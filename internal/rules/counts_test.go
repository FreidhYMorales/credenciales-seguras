package rules

import (
	"testing"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
)

// TestCounts asserts the four state counts printed in the advance document,
// table "Tamanos obtenidos para todas las reglas". The Go port must reproduce
// them digit for digit; this is the port-fidelity gate.
func TestCounts(t *testing.T) {
	// id -> {nfa, subconjuntos, arbol, minimo}
	want := map[string][4]int{
		"P1": {1092, 257, 65, 65},
		"P2": {34, 9, 2, 2},
		"P3": {34, 9, 2, 2},
		"P4": {34, 9, 2, 2},
		"P5": {34, 9, 2, 2},
		"P6": {782, 565, 377, 190},
		"P7": {630, 763, 314, 158},
		"P8": {146, 213, 90, 44},
		"C1": {78, 22, 4, 4},
		"C2": {64, 18, 4, 2},
		"C3": {1404, 321, 65, 65},
		"C4": {2616, 564, 374, 250},
		"C5": {26, 8, 4, 4},
		"C6": {5606, 1276, 256, 256},
		"C*": {136, 33, 12, 8},
	}

	all := All()
	if len(all) != 17 {
		t.Fatalf("registry has %d rules, want 17", len(all))
	}

	seen := 0
	for _, r := range all {
		if !r.TieneER {
			if r.ID != "P0" && r.ID != "C0" {
				t.Errorf("%s has no ER but is not an alphabet check", r.ID)
			}
			continue
		}
		seen++
		w, ok := want[r.ID]
		if !ok {
			t.Errorf("%s has an ER but no expected counts", r.ID)
			continue
		}
		got := r.Counts()
		if got.NFA != w[0] || got.Subset != w[1] || got.Tree != w[2] || got.Min != w[3] {
			t.Errorf("%s counts = %d/%d/%d/%d, want %d/%d/%d/%d",
				r.ID, got.NFA, got.Subset, got.Tree, got.Min, w[0], w[1], w[2], w[3])
		}
	}
	if seen != 15 {
		t.Errorf("%d rules with automaton, want 15", seen)
	}
}

// TestSubsetEqualsTree proves the two construction routes agree, by language
// equivalence rather than by comparing state counts.
func TestSubsetEqualsTree(t *testing.T) {
	for _, r := range All() {
		if !r.TieneER {
			continue
		}
		if !automata.Equivalent(r.Subset, r.Tree) {
			t.Errorf("%s: subset DFA and tree DFA are not equivalent", r.ID)
		}
	}
}

// TestMinEqualsTree proves minimization preserves the language, which is what
// lets the generator walk the product of the minimized DFAs.
func TestMinEqualsTree(t *testing.T) {
	for _, r := range All() {
		if !r.TieneER {
			continue
		}
		if !automata.Equivalent(r.Tree, r.Min) {
			t.Errorf("%s: minimized DFA and tree DFA are not equivalent", r.ID)
		}
	}
}
