package generate

import (
	"strings"
	"testing"

	"github.com/FreidhYMorales/credenciales-seguras/internal/validate"
)

// TestProductBound asserts the reachable product stays within the 1040 states
// the design predicts (65 x 2^4).
func TestProductBound(t *testing.T) {
	p := buildProduct()
	t.Logf("reachable product states: %d", p.Size())
	if p.Size() > 1040 {
		t.Fatalf("product has %d states, want <= 1040", p.Size())
	}
}

// TestPasswordProperty is the property test from the plan: 1000 generated
// values per kind must be 100% accepted.
func TestPasswordProperty(t *testing.T) {
	for seed := int64(0); seed < 1000; seed++ {
		length := 15 + int(seed%50)
		res, err := Password(seed, length)
		if err != nil {
			t.Fatalf("seed %d length %d: %v", seed, length, err)
		}
		if got := len([]rune(res.Valor)); got != length {
			t.Fatalf("seed %d: length = %d, want %d", seed, got, length)
		}
		rep := validate.PasswordReport(res.Valor)
		if !rep.Valido {
			t.Fatalf("seed %d: %q is not valid", seed, res.Valor)
		}
		if len(res.Camino) != length {
			t.Fatalf("seed %d: path has %d steps, want %d", seed, len(res.Camino), length)
		}
	}
}

func TestEmailProperty(t *testing.T) {
	for seed := int64(0); seed < 300; seed++ {
		res, err := Email(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if !validate.EmailReport(res.Valor).Valido {
			t.Fatalf("seed %d: %q is not a valid email", seed, res.Valor)
		}
		if len(res.Camino) != len([]rune(res.Valor)) {
			t.Fatalf("seed %d: path does not cover the value", seed)
		}
	}
}

// TestDeterminism: the same seed reproduces the same value and the same path.
func TestDeterminism(t *testing.T) {
	a, err := Password(2026, 20)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Password(2026, 20)
	if err != nil {
		t.Fatal(err)
	}
	if a.Valor != b.Valor {
		t.Fatalf("same seed gave %q and %q", a.Valor, b.Valor)
	}
	if len(a.Camino) != len(b.Camino) {
		t.Fatal("paths differ in length")
	}
	for i := range a.Camino {
		if a.Camino[i] != b.Camino[i] {
			t.Fatalf("path differs at step %d", i)
		}
	}

	e1, err := Email(4490)
	if err != nil {
		t.Fatal(err)
	}
	e2, err := Email(4490)
	if err != nil {
		t.Fatal(err)
	}
	if e1.Valor != e2.Valor {
		t.Fatalf("same seed gave %q and %q", e1.Valor, e2.Valor)
	}
}

// TestDemoSeeds prints the values the document and the video will use.
func TestDemoSeeds(t *testing.T) {
	p, err := Password(2026, 20)
	if err != nil {
		t.Fatal(err)
	}
	e, err := Email(4490)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("password seed 2026 length 20 -> %s (%s, %v bits)", p.Valor, *p.Nivel, *p.Entropia)
	t.Logf("email    seed 4490          -> %s", e.Valor)

	if strings.Contains(p.Valor, " ") {
		t.Error("generated password contains a space")
	}
	if !validate.PasswordReport(p.Valor).Valido {
		t.Error("demo password is not valid")
	}
}
