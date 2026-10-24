// Package generate produces credentials by walking an automaton, never by
// sampling random strings and filtering them (assignment 7.a.iii).
//
// Passwords walk the product of the minimized DFAs P1 x P2 x P3 x P4 x P5. The
// P1 component is the length counter, so requiring exactly L steps is the same
// as requiring the product to be in an accepting state after L steps. The walk
// only takes transitions from which an accepting state is still reachable in
// the remaining steps, so it never dead-ends.
//
// P6 (repetition), P7 (sequence) and P8 (wordlist) are rejection filters: a draw
// that matches one of them is discarded and redrawn with the same seeded PRNG.
package generate

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
	"github.com/FreidhYMorales/credenciales-seguras/internal/validate"
)

// Step is one move of the recorded walk.
type Step struct {
	Paso    int    `json:"paso"`
	Estado  string `json:"estado"`
	Simbolo string `json:"simbolo"`
	Clase   string `json:"clase"`
}

// Result is the payload of POST /api/generar.
type Result struct {
	Valor    string   `json:"valor"`
	Tipo     string   `json:"tipo"`
	Longitud int      `json:"longitud"`
	Nivel    *string  `json:"nivel"`
	Entropia *float64 `json:"entropia"`
	Semilla  int64    `json:"semilla"`
	Camino   []Step   `json:"camino"`
}

// ErrExhausted is returned when rejection sampling cannot find an acceptable
// value; it should not happen with the shipped rules.
var ErrExhausted = errors.New("no se pudo generar una credencial válida")

const (
	minPasswordLen = 15
	maxPasswordLen = 64
	maxRetries     = 200
)

// Password walks the product automaton and returns a password of exactly
// length characters, or an error.
func Password(seed int64, length int) (Result, error) {
	if length == 0 {
		length = 20
	}
	if length < minPasswordLen || length > maxPasswordLen {
		return Result{}, errors.New("la longitud debe estar entre 15 y 64")
	}

	p := buildProduct()
	rng := rand.New(rand.NewSource(seed))

	for attempt := 0; attempt < maxRetries; attempt++ {
		value, path, ok := p.walk(rng, length)
		if !ok {
			continue
		}
		if rejected(value) {
			continue
		}
		rep := validate.PasswordReport(value)
		if !rep.Valido {
			continue
		}
		return Result{
			Valor:    value,
			Tipo:     string(validate.Password),
			Longitud: rep.Longitud,
			Nivel:    rep.Nivel,
			Entropia: rep.Entropia,
			Semilla:  seed,
			Camino:   path,
		}, nil
	}
	return Result{}, ErrExhausted
}

// rejected reports whether the value matches a rejection filter (P6-P8).
func rejected(s string) bool {
	for _, id := range []string{"P6", "P7", "P8"} {
		if rules.Get(id).Accepts(s) {
			return true
		}
	}
	return false
}

// Email walks the minimized DFA of C* with the same reachability guidance as the
// password walk, then reject-samples against the full validator, which covers
// C2-C6.
//
// A purely uniform walk over C* never finds the '@': the guided walk only takes
// transitions from which an accepting state is still reachable in the remaining
// steps, which is what forces the local part, the '@', the labels and the TLD to
// appear in order.
func Email(seed int64) (Result, error) {
	const (
		minLen = 12
		maxLen = 32
	)
	c := rules.Get("C*")
	rng := rand.New(rand.NewSource(seed))
	reach := reachTable(c.Min, maxLen)

	for attempt := 0; attempt < maxRetries; attempt++ {
		length := minLen + rng.Intn(maxLen-minLen+1)
		if !reach[c.Min.Start][length] {
			continue
		}
		value, path, ok := walkDFA(c.Min, reach, rng, length, true)
		if !ok {
			continue
		}
		// Quality filter: course material shows conservative addresses, so the
		// atext specials are reject-sampled away even though the grammar allows
		// them. The API still accepts them on the validation side.
		if strings.ContainsAny(value, "!#$%&'*+/=?^_`{|}~") {
			continue
		}
		if !validate.EmailReport(value).Valido {
			continue
		}
		return Result{
			Valor:    value,
			Tipo:     string(validate.Email),
			Longitud: len([]rune(value)),
			Semilla:  seed,
			Camino:   path,
		}, nil
	}
	return Result{}, ErrExhausted
}

// reachTable returns reach[state][r] = an accepting state is reachable from
// state in exactly r steps.
func reachTable(d *automata.DFA, max int) [][]bool {
	reach := make([][]bool, d.N)
	for i := range reach {
		reach[i] = make([]bool, max+1)
		reach[i][0] = d.Accept[i]
	}
	for r := 1; r <= max; r++ {
		for i := 0; i < d.N; i++ {
			ok := false
			for a := range d.Atoms {
				if t := d.Delta[i][a]; t >= 0 && reach[t][r-1] {
					ok = true
					break
				}
			}
			reach[i][r] = ok
		}
	}
	return reach
}

// walkDFA walks exactly length symbols, only through transitions that keep an
// accepting state reachable in the remaining steps. When weighted is true the
// atom is drawn proportionally to its size, which makes emails look like emails
// (letters dominate) without weakening the guarantee.
func walkDFA(d *automata.DFA, reach [][]bool, rng *rand.Rand, length int, weighted bool) (string, []Step, bool) {
	if !reach[d.Start][length] {
		return "", nil, false
	}
	id := d.Start
	out := make([]rune, 0, length)
	path := make([]Step, 0, length)

	for k := 0; k < length; k++ {
		remaining := length - k - 1
		var cands []int
		for a := range d.Atoms {
			if t := d.Delta[id][a]; t >= 0 && reach[t][remaining] {
				cands = append(cands, a)
			}
		}
		if len(cands) == 0 {
			return "", nil, false
		}
		a := pickAtom(d, cands, rng, weighted)
		atom := d.Atoms[a]
		r := atom.Chars[rng.Intn(len(atom.Chars))]
		out = append(out, r)
		path = append(path, Step{
			Paso:    k,
			Estado:  "q" + strconv.Itoa(id),
			Simbolo: string(r),
			Clase:   classOf(r),
		})
		id = d.Delta[id][a]
	}
	if !d.Accept[id] {
		return "", nil, false
	}
	return string(out), path, true
}

func pickAtom(d *automata.DFA, cands []int, rng *rand.Rand, weighted bool) int {
	if !weighted {
		return cands[rng.Intn(len(cands))]
	}
	total := 0
	for _, a := range cands {
		w := len(d.Atoms[a].Chars)
		total += w * w
	}
	n := rng.Intn(total)
	for _, a := range cands {
		n -= len(d.Atoms[a].Chars) * len(d.Atoms[a].Chars)
		if n < 0 {
			return a
		}
	}
	return cands[len(cands)-1]
}

// classOf returns the display label of the symbol's class.
func classOf(r rune) string {
	switch {
	case r >= 'A' && r <= 'Z':
		return "L"
	case r >= 'a' && r <= 'z':
		return "l"
	case r >= '0' && r <= '9':
		return "N"
	default:
		return "S"
	}
}
