// Package rules defines the alphabets, the labels and the regular expressions of
// every rule, and builds their automata.
//
// Rule ids are frozen: P0..P8 (passwords), C0..C6 (emails) and C* (whole email).
// P0 and C0 are alphabet set checks and have no automaton; the remaining 15 do.
package rules

import (
	"sync"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
)

// Alphabet of the passwords: the 94 printable ASCII symbols. The space (0x20)
// is stripped before validation, so it is not part of the alphabet.
var (
	Upper     = runeRange('A', 'Z')
	Lower     = runeRange('a', 'z')
	Digit     = runeRange('0', '9')
	Printable = runeRange(0x21, 0x7E)
	Special   = minus(Printable, Upper, Lower, Digit)
)

// AtextSpecial is the RFC 5322 atext set excluding the hyphen, which has its own
// label H.
var AtextSpecial = []rune("!#$%&'*+/=?^_`{|}~")

// Labels shared by the expressions below.
var (
	LeafUpper    = automata.Leaf(Upper, "L")
	LeafLower    = automata.Leaf(Lower, "l")
	LeafDigit    = automata.Leaf(Digit, "N")
	LeafSpecial  = automata.Leaf(Special, "S")
	SigmaPwd     = automata.Alt(LeafUpper, LeafLower, LeafDigit, LeafSpecial)
	starSigmaPwd = automata.Star(SigmaPwd)
)

// Password expressions.
var (
	PwdLength   = automata.Rep(SigmaPwd, 15, 64)
	PwdUpper    = contains(LeafUpper)
	PwdLower    = contains(LeafLower)
	PwdDigit    = contains(LeafDigit)
	PwdSpecial  = contains(LeafSpecial)
	PwdRepeat   = contains(triples())
	PwdSequence = contains(sequences())
	PwdWordlist = contains(wordlist())
)

func contains(x automata.Node) automata.Node {
	return automata.Cat(starSigmaPwd, x, starSigmaPwd)
}

func triples() automata.Node {
	parts := make([]automata.Node, 0, len(Printable))
	for _, c := range Printable {
		one := automata.Leaf([]rune{c}, string(c))
		parts = append(parts, automata.Cat(one, one, one))
	}
	return automata.Alt(parts...)
}

func sequences() automata.Node {
	bases := [][]rune{Lower, Digit}
	var parts []automata.Node
	for _, base := range bases {
		for i := 0; i+4 <= len(base); i++ {
			w := base[i : i+4]
			parts = append(parts, word(w))
			parts = append(parts, word(reverse(w)))
		}
	}
	return automata.Alt(parts...)
}

func wordlist() automata.Node {
	words := []string{"password", "contrasena", "admin", "qwerty", "iloveyou", "welcome", "letmein"}
	parts := make([]automata.Node, 0, len(words))
	for _, w := range words {
		parts = append(parts, word([]rune(w)))
	}
	return automata.Alt(parts...)
}

// word matches the letters of w case-insensitively.
func word(w []rune) automata.Node {
	parts := make([]automata.Node, 0, len(w))
	for _, c := range w {
		parts = append(parts, caseInsensitive(c))
	}
	return automata.Cat(parts...)
}

func caseInsensitive(c rune) automata.Node {
	chars := []rune{c}
	if c >= 'a' && c <= 'z' {
		chars = append(chars, c-'a'+'A')
	}
	return automata.Leaf(chars, string(c))
}

// Email expressions.
var (
	LeafLetter = automata.Leaf(concat(Upper, Lower), "A")
	LeafHyphen = automata.Leaf([]rune{'-'}, "H")
	LeafAtext  = automata.Leaf(AtextSpecial, "E")
	LeafDot    = automata.Leaf([]rune{'.'}, "P")
	LeafAt     = automata.Leaf([]rune{'@'}, "@")
	LeafEmailN = automata.Leaf(Digit, "N")

	atext         = automata.Alt(LeafLetter, LeafEmailN, LeafHyphen, LeafAtext)
	noAt          = automata.Alt(LeafLetter, LeafEmailN, LeafHyphen, LeafAtext, LeafDot)
	label         = automata.Cat(automata.Alt(LeafLetter, LeafEmailN), automata.Opt(automata.Cat(automata.Star(automata.Alt(LeafLetter, LeafEmailN, LeafHyphen)), automata.Alt(LeafLetter, LeafEmailN))))
	label63       = automata.Cat(automata.Alt(LeafLetter, LeafEmailN), automata.Opt(automata.Cat(automata.Rep(automata.Alt(LeafLetter, LeafEmailN, LeafHyphen), 0, 61), automata.Alt(LeafLetter, LeafEmailN))))
	EmailStruct   = automata.Cat(automata.Plus(noAt), LeafAt, automata.Plus(noAt))
	EmailLocal    = automata.Cat(automata.Plus(atext), automata.Star(automata.Cat(LeafDot, automata.Plus(atext))))
	EmailLocalLen = automata.Rep(noAt, 1, 64)
	EmailLabels   = automata.Cat(label63, automata.Plus(automata.Cat(LeafDot, label63)))
	EmailTLD      = automata.Cat(automata.Star(automata.Alt(LeafLetter, LeafEmailN, LeafHyphen, LeafDot)), LeafDot, LeafLetter, automata.Plus(LeafLetter))
	EmailDomLen   = automata.Rep(noAt, 1, 255)
	EmailFull     = automata.Cat(
		automata.Plus(atext), automata.Star(automata.Cat(LeafDot, automata.Plus(atext))),
		LeafAt,
		automata.Plus(automata.Cat(label, LeafDot)),
		LeafLetter, automata.Plus(LeafLetter),
	)
)

// Rule is one row of the rules table.
type Rule struct {
	ID      string
	Nombre  string
	Tipo    string // "accept" or "reject"
	ER      string // formal expression for display; empty when TieneER is false
	TieneER bool
	Root    automata.Node
	*automata.Compiled
}

// Stats exposes the four state counts of the rule.
func (r *Rule) Counts() automata.Stats { return r.Compiled.Stats() }

var (
	once     sync.Once
	registry []*Rule
	byID     map[string]*Rule
)

// All returns the 17 rules in id order. The automata are built once.
func All() []*Rule {
	once.Do(build)
	return registry
}

// Get returns the rule with the given id, or nil.
func Get(id string) *Rule {
	once.Do(build)
	return byID[id]
}

// PasswordRules and EmailRules return the rules of each kind, in id order.
func PasswordRules() []*Rule {
	once.Do(build)
	out := make([]*Rule, 0, 9)
	for _, r := range registry {
		if r.ID[0] == 'P' {
			out = append(out, r)
		}
	}
	return out
}

func EmailRules() []*Rule {
	once.Do(build)
	out := make([]*Rule, 0, 7)
	for _, r := range registry {
		// C* also starts with C but is the whole-email rule, not one of C0..C6.
		if r.ID == "C*" {
			continue
		}
		if r.ID[0] == 'C' {
			out = append(out, r)
		}
	}
	return out
}

func build() {
	defs := []struct {
		id, nombre, tipo, er string
		root                 automata.Node
	}{
		{"P0", "Alfabeto", "accept", "", nil},
		{"P1", "Longitud 15-64", "accept", "(L|l|N|S){15,64}", PwdLength},
		{"P2", "Mayúscula", "accept", "(L|l|N|S)* L (L|l|N|S)*", PwdUpper},
		{"P3", "Minúscula", "accept", "(L|l|N|S)* l (L|l|N|S)*", PwdLower},
		{"P4", "Dígito", "accept", "(L|l|N|S)* N (L|l|N|S)*", PwdDigit},
		{"P5", "Especial", "accept", "(L|l|N|S)* S (L|l|N|S)*", PwdSpecial},
		{"P6", "Repetición (rechazo)", "reject", "(L|l|N|S)* (c c c) (L|l|N|S)*, c en Sigma", PwdRepeat},
		{"P7", "Secuencia (rechazo)", "reject", "(L|l|N|S)* (a b c d) (L|l|N|S)*, 4 en orden", PwdSequence},
		{"P8", "Palabra prohibida (rechazo)", "reject", "(L|l|N|S)* (palabra) (L|l|N|S)*", PwdWordlist},
		{"C0", "Alfabeto", "accept", "", nil},
		{"C1", "Estructura", "accept", "(A|N|H|E|P)+ @ (A|N|H|E|P)+", EmailStruct},
		{"C2", "Parte local", "accept", "atext+ (. atext+)*", EmailLocal},
		{"C3", "Longitud local", "accept", "(A|N|H|E|P){1,64}", EmailLocalLen},
		{"C4", "Etiquetas del dominio", "accept", "label63 (. label63)+", EmailLabels},
		{"C5", "TLD", "accept", "(A|N|H|P)* . A A+", EmailTLD},
		{"C6", "Longitud del dominio", "accept", "(A|N|H|E|P){1,255}", EmailDomLen},
		{"C*", "Correo completo", "accept", "local @ dominio . TLD", EmailFull},
	}

	registry = make([]*Rule, 0, len(defs))
	byID = make(map[string]*Rule, len(defs))
	for _, d := range defs {
		r := &Rule{ID: d.id, Nombre: d.nombre, Tipo: d.tipo, ER: d.er, TieneER: d.root != nil}
		if d.root != nil {
			r.Compiled = automata.Compile(d.root)
		}
		registry = append(registry, r)
		byID[d.id] = r
	}
}

func runeRange(lo, hi rune) []rune {
	out := make([]rune, 0, hi-lo+1)
	for r := lo; r <= hi; r++ {
		out = append(out, r)
	}
	return out
}

func minus(sets ...[]rune) []rune {
	base := sets[0]
	var out []rune
	for _, r := range base {
		keep := true
		for _, s := range sets[1:] {
			for _, x := range s {
				if x == r {
					keep = false
					break
				}
			}
		}
		if keep {
			out = append(out, r)
		}
	}
	return out
}

func concat(xs ...[]rune) []rune {
	var out []rune
	for _, x := range xs {
		out = append(out, x...)
	}
	return out
}

func reverse(w []rune) []rune {
	out := make([]rune, len(w))
	for i := range w {
		out[len(w)-1-i] = w[i]
	}
	return out
}
