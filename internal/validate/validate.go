// Package validate produces the per-rule lexical report and the security level.
//
// The report always contains every rule of the requested kind. A rule that was
// skipped because an earlier rule already decided the verdict is returned with
// Evaluado=false, so the table in the UI is never incomplete.
package validate

import (
	"fmt"
	"math"
	"strings"

	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
)

// Kind is the kind of credential.
type Kind string

const (
	Password Kind = "password"
	Email    Kind = "email"
)

// RuleResult is one row of the report.
type RuleResult struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	Tipo     string `json:"tipo"`
	OK       bool   `json:"ok"`
	Evaluado bool   `json:"evaluado"`
	Mensaje  string `json:"mensaje"`
	Posicion *int   `json:"posicion"` // 0-based, only for P0 and C0
}

// Report is the response payload of POST /api/validar.
type Report struct {
	Valido      bool         `json:"valido"`
	Normalizado string       `json:"normalizado"`
	Longitud    int          `json:"longitud"`
	Nivel       *string      `json:"nivel"`
	Entropia    *float64     `json:"entropia"`
	Reglas      []RuleResult `json:"reglas"`
}

// Normalize removes the spaces, which are not part of either alphabet.
func Normalize(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

// Validate dispatches on the kind.
func Validate(kind Kind, raw string) Report {
	switch kind {
	case Password:
		return PasswordReport(raw)
	case Email:
		return EmailReport(raw)
	}
	return Report{}
}

// PasswordReport validates a password against P0-P8.
func PasswordReport(raw string) Report {
	s := Normalize(raw)
	runes := []rune(s)
	rep := Report{Normalizado: s, Longitud: len(runes)}
	passwordRules := rules.PasswordRules()

	// P0: alphabet check, no automaton.
	badIdx := -1
	for i, r := range runes {
		if !containsRune(rules.Printable, r) {
			badIdx = i
			break
		}
	}
	p0 := ruleResult(rules.Get("P0"), badIdx < 0)
	if badIdx >= 0 {
		p0.Posicion = &badIdx
		p0.Mensaje = fmt.Sprintf("símbolo no permitido en la posición %d", badIdx+1)
	}
	rep.Reglas = append(rep.Reglas, p0)
	if badIdx >= 0 {
		// Nothing else can be evaluated on a string outside the alphabet.
		for _, r := range passwordRules[1:] {
			rep.Reglas = append(rep.Reglas, skipped(r))
		}
		rep.Valido = false
		rep.Nivel = levelPointer("Baja")
		return rep
	}

	for _, r := range passwordRules[1:] {
		hit := r.Accepts(s)
		res := ruleResult(r, hit)
		switch r.ID {
		case "P1":
			n := len(runes)
			res.Mensaje = lengthMessage(n)
		case "P2", "P3", "P4", "P5":
			if !hit {
				res.Mensaje = map[string]string{
					"P2": "falta mayúscula",
					"P3": "falta minúscula",
					"P4": "falta dígito",
					"P5": "falta carácter especial",
				}[r.ID]
			}
		default: // P6-P8 are rejection filters: the automaton matches the bad case.
			res.OK = !hit
			if hit {
				res.Mensaje = map[string]string{
					"P6": "3 o más caracteres idénticos consecutivos",
					"P7": "secuencia ascendente o descendente de 4",
					"P8": "contiene una palabra prohibida",
				}[r.ID]
			}
		}
		rep.Reglas = append(rep.Reglas, res)
	}

	rep.Valido = allOK(rep.Reglas)
	if rep.Valido {
		rep.Nivel = levelPointer(Level(rep.Longitud))
		bits := Entropy(rep.Longitud)
		rep.Entropia = &bits
	} else {
		rep.Nivel = levelPointer("Baja")
	}
	return rep
}

// EmailReport validates an email against C0-C6.
func EmailReport(raw string) Report {
	s := Normalize(raw)
	runes := []rune(s)
	rep := Report{Normalizado: s, Longitud: len(runes)}
	emailRules := rules.EmailRules()

	badIdx := -1
	for i, r := range runes {
		if !containsRune(emailAlphabet, r) {
			badIdx = i
			break
		}
	}
	c0 := ruleResult(rules.Get("C0"), badIdx < 0)
	if badIdx >= 0 {
		c0.Posicion = &badIdx
		c0.Mensaje = fmt.Sprintf("símbolo no permitido en la posición %d", badIdx+1)
	}
	rep.Reglas = append(rep.Reglas, c0)
	if badIdx >= 0 {
		for _, r := range emailRules[1:] {
			rep.Reglas = append(rep.Reglas, skipped(r))
		}
		rep.Valido = false
		return rep
	}

	c1 := rules.Get("C1")
	ok1 := c1.Accepts(s)
	res1 := ruleResult(c1, ok1)
	if !ok1 {
		res1.Mensaje = fmt.Sprintf("%d arrobas; se espera local@dominio", strings.Count(s, "@"))
	}
	rep.Reglas = append(rep.Reglas, res1)
	if !ok1 {
		for _, r := range emailRules[2:] {
			rep.Reglas = append(rep.Reglas, skipped(r))
		}
		rep.Valido = false
		return rep
	}

	local, domain, _ := strings.Cut(s, "@")
	for _, r := range emailRules[2:] {
		part := local
		if r.ID == "C4" || r.ID == "C5" || r.ID == "C6" {
			part = domain
		}
		ok := r.Accepts(part)
		res := ruleResult(r, ok)
		if !ok {
			res.Mensaje = map[string]string{
				"C2": "dot-atom inválido (punto inicial, final o doble)",
				"C3": "parte local mayor a 64",
				"C4": "etiqueta inválida (guion en el borde, vacía o mayor a 63) o sin punto",
				"C5": "el TLD debe ser solo letras y tener 2 o más",
				"C6": "dominio mayor a 255",
			}[r.ID]
		}
		rep.Reglas = append(rep.Reglas, res)
	}

	rep.Valido = allOK(rep.Reglas)
	return rep
}

// Level applies Annex C. Only called for valid passwords; a valid password has
// between 15 and 64 characters, so the two remaining bands are exhaustive.
func Level(n int) string {
	if n >= 20 {
		return "Fuerte"
	}
	return "Media"
}

// Entropy is the gross estimate of Annex C: n * log2(94), in bits.
func Entropy(n int) float64 {
	return math.Round(float64(n) * math.Log2(94))
}

func lengthMessage(n int) string {
	switch {
	case n < 15:
		return fmt.Sprintf("longitud = %d (faltan %d)", n, 15-n)
	case n > 64:
		return fmt.Sprintf("longitud = %d (sobran %d)", n, n-64)
	default:
		return fmt.Sprintf("longitud = %d", n)
	}
}

func ruleResult(r *rules.Rule, ok bool) RuleResult {
	return RuleResult{
		ID:       r.ID,
		Nombre:   r.Nombre,
		Tipo:     r.Tipo,
		OK:       ok,
		Evaluado: true,
	}
}

func skipped(r *rules.Rule) RuleResult {
	return RuleResult{ID: r.ID, Nombre: r.Nombre, Tipo: r.Tipo, OK: false, Evaluado: false}
}

func allOK(rs []RuleResult) bool {
	for _, r := range rs {
		if !r.OK {
			return false
		}
	}
	return true
}

func levelPointer(s string) *string { return &s }

func containsRune(xs []rune, r rune) bool {
	for _, x := range xs {
		if x == r {
			return true
		}
	}
	return false
}

// emailAlphabet is the 83-symbol alphabet of the email rules.
var emailAlphabet = func() []rune {
	out := make([]rune, 0, 83)
	out = append(out, rules.Upper...)
	out = append(out, rules.Lower...)
	out = append(out, rules.Digit...)
	out = append(out, rules.AtextSpecial...)
	out = append(out, '.', '@')
	return out
}()
