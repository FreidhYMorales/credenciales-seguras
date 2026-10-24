package rules

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

// samples are strings each rule accepts, used to seed the mutation fuzzer so it
// exercises the accepting boundary and not only the rejection path.
var samples = map[string]string{
	"P1": "abcdefghijklmno",
	"P2": "A",
	"P3": "a",
	"P4": "1",
	"P5": "!",
	"P6": "aaa",
	"P7": "abcd",
	"P8": "password",
	"C1": "a@b",
	"C2": "a.b",
	"C3": "a",
	"C4": "a.b",
	"C5": ".co",
	"C6": "a",
	"C*": "a@b.co",
}

func poolFor(id string) []rune {
	if id[0] == 'C' {
		return []rune("abzAZ09.-_+@")
	}
	return []rune("abzAZ09!@#.-_")
}

func maxLenFor(id string) int {
	if id[0] == 'C' {
		return 300
	}
	return 70
}

func mutate(rng *rand.Rand, s string, pool []rune) string {
	b := []rune(s)
	for n := rng.Intn(4); n > 0; n-- {
		switch rng.Intn(3) {
		case 0:
			pos := rng.Intn(len(b) + 1)
			b = append(b[:pos], append([]rune{pool[rng.Intn(len(pool))]}, b[pos:]...)...)
		case 1:
			if len(b) > 0 {
				pos := rng.Intn(len(b))
				b = append(b[:pos], b[pos+1:]...)
			}
		case 2:
			if len(b) > 0 {
				b[rng.Intn(len(b))] = pool[rng.Intn(len(pool))]
			}
		}
	}
	return string(b)
}

func randomString(rng *rand.Rand, pool []rune, max int) string {
	n := rng.Intn(max + 1)
	b := make([]rune, n)
	for i := range b {
		b[i] = pool[rng.Intn(len(pool))]
	}
	return string(b)
}

// TestRE2Oracle is the independent check: the automaton and Go's regexp (RE2),
// compiled from the same AST through toRE2, must agree on every generated
// string. 1500 strings per rule, as stated in the plan.
func TestRE2Oracle(t *testing.T) {
	const perRule = 1500
	rng := rand.New(rand.NewSource(12345))

	for _, r := range All() {
		if !r.TieneER {
			continue
		}
		re, err := regexp.Compile(r.RE2())
		if err != nil {
			t.Fatalf("%s: RE2 does not compile: %v\n%s", r.ID, err, r.RE2())
		}
		pool := poolFor(r.ID)
		max := maxLenFor(r.ID)
		sample := samples[r.ID]
		bad := 0
		for i := 0; i < perRule; i++ {
			var s string
			if i%2 == 0 {
				s = mutate(rng, sample, pool)
			} else {
				s = randomString(rng, pool, max)
			}
			got := r.Accepts(s)
			want := re.MatchString(s)
			if got != want {
				bad++
				if bad <= 5 {
					t.Errorf("%s: %q automata=%v re2=%v", r.ID, s, got, want)
				}
			}
		}
		if bad > 0 {
			t.Errorf("%s: %d/%d discrepancies", r.ID, bad, perRule)
		}
	}
}

// TestAcceptReject is the table from the document's test annex: every rule has
// at least one accepted and one rejected string. For the reject rules (P6-P8)
// "accepted" means the automaton recognises the insecure pattern.
func TestAcceptReject(t *testing.T) {
	cases := []struct {
		id     string
		accept []string
		reject []string
	}{
		{"P1", []string{"abcdefghijklmno", strings.Repeat("a", 64)}, []string{strings.Repeat("a", 14), strings.Repeat("a", 65), ""}},
		{"P2", []string{"A", "aA"}, []string{"a", "1"}},
		{"P3", []string{"a", "Aa"}, []string{"A", "1"}},
		{"P4", []string{"1", "a1"}, []string{"a", "!"}},
		{"P5", []string{"!", "a!"}, []string{"a", "1"}},
		{"P6", []string{"aaa", "x111y"}, []string{"aab", "abab"}},
		{"P7", []string{"abcd", "1234", "dcba"}, []string{"abce", "12ab"}},
		{"P8", []string{"password", "Qwerty", "contrasena"}, []string{"passwerd", "adm1n"}},
		{"C1", []string{"a@b", "us.er@do.main"}, []string{"ab", "a@@b", "@b"}},
		{"C2", []string{"a.b", "a"}, []string{".a", "a..b"}},
		{"C3", []string{"a", "abc"}, []string{""}},
		{"C4", []string{"a.b", "a.b.c"}, []string{"a", "a..b"}},
		{"C5", []string{".co", ".com"}, []string{".c", "com"}},
		{"C6", []string{"a", "example.com"}, []string{""}},
		{"C*", []string{"a@b.co", "ana.perez@correo.com"}, []string{"a@b", "ana@correo", "@correo.com"}},
	}

	for _, c := range cases {
		r := Get(c.id)
		if r == nil || !r.TieneER {
			t.Fatalf("%s: rule missing or without automaton", c.id)
		}
		for _, s := range c.accept {
			if !r.Accepts(s) {
				t.Errorf("%s: %q should be accepted", c.id, s)
			}
		}
		for _, s := range c.reject {
			if r.Accepts(s) {
				t.Errorf("%s: %q should be rejected", c.id, s)
			}
		}
	}
}
