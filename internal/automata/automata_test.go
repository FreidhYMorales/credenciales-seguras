package automata_test

import (
	"regexp"
	"testing"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
)

// TestToRE2EscapesEverySymbol is the guard the alphabet needs: the 94 printable
// symbols include '\', ']', '^', '-' and '[', all of which are special inside a
// character class. A naive [...] would either fail to compile or silently match
// a different language.
func TestToRE2EscapesEverySymbol(t *testing.T) {
	for r := rune(0x21); r <= 0x7e; r++ {
		n := automata.Leaf([]rune{r}, string(r))
		expr := automata.ToRE2(n)
		re, err := regexp.Compile(expr)
		if err != nil {
			t.Errorf("%q: RE2 does not compile: %v (%s)", r, err, expr)
			continue
		}
		if !re.MatchString(string(r)) {
			t.Errorf("%q: its own translation does not match it (%s)", r, expr)
		}
		// The class must be exactly one symbol: a neighbour must not match.
		for _, other := range []rune{'a', 'A', '0', '!', '\\', ']', '^', '-', '['} {
			if other == r {
				continue
			}
			if re.MatchString(string(other)) {
				t.Errorf("%q: translation %s also matches %q", r, expr, other)
			}
		}
	}
}

// TestToRE2EscapingIsNeeded proves the test above would catch a naive class: a
// hand-written unescaped class over the same characters is invalid RE2.
func TestToRE2EscapingIsNeeded(t *testing.T) {
	if _, err := regexp.Compile(`\A(?:[\])\z`); err == nil {
		t.Fatal("an unescaped ']' inside a class should not compile")
	}
}

// TestMinimizePreservesLanguage checks the minimization on expressions with a
// non-trivial dead state: the minimal DFA must accept exactly what the tree DFA
// accepts.
func TestMinimizePreservesLanguage(t *testing.T) {
	cases := []struct {
		name   string
		root   automata.Node
		accept []string
		reject []string
	}{
		{
			name: "optional suffix",
			root: automata.Cat(
				automata.Leaf([]rune{'a', 'b'}, "ab"),
				automata.Opt(automata.Leaf([]rune{'c'}, "c")),
			),
			accept: []string{"a", "b", "ac", "bc"},
			reject: []string{"", "c", "abc", "aa"},
		},
		{
			name: "star of an alternation",
			root: automata.Star(automata.Alt(
				automata.Leaf([]rune{'a'}, "a"),
				automata.Leaf([]rune{'b'}, "b"),
			)),
			accept: []string{"", "a", "b", "abba", "bbbb"},
			reject: []string{"c", "abx", "a b"},
		},
		{
			name:   "bounded repetition",
			root:   automata.Rep(automata.Leaf([]rune{'x'}, "x"), 2, 4),
			accept: []string{"xx", "xxx", "xxxx"},
			reject: []string{"", "x", "xxxxx", "xy"},
		},
	}

	for _, c := range cases {
		comp := automata.Compile(c.root)
		if !automata.Equivalent(comp.Tree, comp.Min) {
			t.Errorf("%s: minimized DFA is not equivalent to the tree DFA", c.name)
		}
		if !automata.Equivalent(comp.Subset, comp.Min) {
			t.Errorf("%s: minimized DFA is not equivalent to the subset DFA", c.name)
		}
		for _, s := range c.accept {
			if !comp.Min.Accepts([]rune(s)) {
				t.Errorf("%s: %q should be accepted by the minimal DFA", c.name, s)
			}
		}
		for _, s := range c.reject {
			if comp.Min.Accepts([]rune(s)) {
				t.Errorf("%s: %q should be rejected by the minimal DFA", c.name, s)
			}
		}
	}
}

// TestMinimizeDropsDeadState pins the convention the 65-state figure depends on:
// the minimal DFA is partial, so a missing transition means reject. For "ab" the
// complete DFA has four states (q0, q1, q2 and the dead state); the minimal one
// keeps three.
func TestMinimizeDropsDeadState(t *testing.T) {
	root := automata.Cat(
		automata.Leaf([]rune{'a'}, "a"),
		automata.Leaf([]rune{'b'}, "b"),
	)
	comp := automata.Compile(root)
	if comp.Min.N != 3 {
		t.Fatalf("minimal DFA has %d states, want 3 (the dead state is dropped)", comp.Min.N)
	}
	if comp.Min.Accepts([]rune("ax")) {
		t.Error("a missing transition must reject")
	}
	if !comp.Min.Accepts([]rune("ab")) {
		t.Error("ab should be accepted")
	}
}

// TestSimulateTrace checks the trace and the failure index the explorer shows.
func TestSimulateTrace(t *testing.T) {
	root := automata.Cat(
		automata.Leaf([]rune{'a', 'b'}, "ab"),
		automata.Leaf([]rune{'1', '2'}, "12"),
	)
	d := automata.Compile(root).Min

	acepta, traza, fallo := automata.Simulate(d, []rune("a1"))
	if !acepta || fallo != -1 {
		t.Fatalf("a1: acepta=%v fallo=%d, want true/-1", acepta, fallo)
	}
	if len(traza) != 2 {
		t.Fatalf("trace has %d steps, want 2", len(traza))
	}
	if traza[0].Paso != 0 || traza[0].Desde != d.Start || traza[0].Simbolo != 'a' {
		t.Errorf("first step = %+v", traza[0])
	}

	acepta, traza, fallo = automata.Simulate(d, []rune("ax"))
	if acepta {
		t.Error("ax should be rejected")
	}
	if fallo != 1 {
		t.Errorf("fallo = %d, want 1 (the 'x')", fallo)
	}
	if len(traza) != 2 || traza[1].Hacia != -1 {
		t.Errorf("trace should stop on the failing symbol: %+v", traza)
	}
}

// TestAtomsPartition checks the property the subset construction relies on:
// every symbol of the universe belongs to exactly one atom.
func TestAtomsPartition(t *testing.T) {
	root := automata.Alt(
		automata.Leaf([]rune{'a', 'b', 'c'}, "abc"),
		automata.Leaf([]rune{'b', 'c', 'd'}, "bcd"),
	)
	atoms, c2a := automata.Atoms(root)

	seen := map[rune]int{}
	for i, at := range atoms {
		if len(at.Chars) == 0 {
			t.Errorf("atom %d is empty", i)
		}
		for _, r := range at.Chars {
			seen[r]++
		}
	}
	for r, n := range seen {
		if n != 1 {
			t.Errorf("%q appears in %d atoms, want 1", r, n)
		}
		if _, ok := c2a[r]; !ok {
			t.Errorf("%q is in an atom but missing from the char-to-atom map", r)
		}
	}
	// 'a' and 'd' belong to different atoms: only 'b' and 'c' are in both leaves.
	if c2a['a'] == c2a['b'] {
		t.Error("'a' and 'b' should not share an atom")
	}
	if c2a['b'] != c2a['c'] {
		t.Error("'b' and 'c' belong to the same leaves and should share an atom")
	}
}
