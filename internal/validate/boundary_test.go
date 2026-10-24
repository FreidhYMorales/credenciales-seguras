package validate

import (
	"strings"
	"testing"
)

// pwdOfLength builds a password of exactly n characters that satisfies every
// acceptance rule: it repeats the safe block "Xk9$", which has an uppercase, a
// lowercase, a digit and a special, no triple repetition and no four-character
// run in the lowercase or digit ranges.
func pwdOfLength(n int) string {
	const block = "Xk9$"
	var b strings.Builder
	for b.Len() < n {
		b.WriteString(block)
	}
	return b.String()[:n]
}

// TestPasswordLengthBoundaries pins 15-64 inclusive: 14 and 65 must fail P1 and
// only P1, while 15 and 64 must be valid.
func TestPasswordLengthBoundaries(t *testing.T) {
	cases := []struct {
		n     int
		valid bool
	}{
		{14, false},
		{15, true},
		{16, true},
		{63, true},
		{64, true},
		{65, false},
	}
	for _, c := range cases {
		valor := pwdOfLength(c.n)
		if got := len([]rune(valor)); got != c.n {
			t.Fatalf("helper produced %d characters, want %d", got, c.n)
		}
		rep := PasswordReport(valor)
		if rep.Valido != c.valid {
			t.Errorf("longitud %d: valido = %v, want %v (%s)", c.n, rep.Valido, c.valid, failureSummary(rep))
		}
		if rep.Longitud != c.n {
			t.Errorf("longitud %d: report says %d", c.n, rep.Longitud)
		}
		if c.valid {
			continue
		}
		// The only failure must be P1: the rest of the value is sound.
		for _, r := range rep.Reglas {
			if r.ID != "P1" && r.Evaluado && !r.OK {
				t.Errorf("longitud %d: %s also fails (%s), so the case is not isolating P1", c.n, r.ID, r.Mensaje)
			}
		}
	}
}

// TestPasswordLevelBoundaries pins the Annex C split at 20 characters: 19 is
// Media, 20 is Fuerte.
func TestPasswordLevelBoundaries(t *testing.T) {
	cases := []struct {
		n     int
		nivel string
	}{
		{15, "Media"},
		{19, "Media"},
		{20, "Fuerte"},
		{64, "Fuerte"},
	}
	for _, c := range cases {
		rep := PasswordReport(pwdOfLength(c.n))
		if !rep.Valido {
			t.Errorf("longitud %d: should be valid (%s)", c.n, failureSummary(rep))
			continue
		}
		if rep.Nivel == nil || *rep.Nivel != c.nivel {
			t.Errorf("longitud %d: nivel = %v, want %s", c.n, rep.Nivel, c.nivel)
		}
	}
}

// TestLongBlocklistedPasswordIsBaja covers the Annex C edge case the document
// calls out: a password long enough for Fuerte that trips a rejection filter
// drops to Baja, not Media.
func TestLongBlocklistedPasswordIsBaja(t *testing.T) {
	const valor = "Xk9$mTq2#vLw8@adminX" // exactly 20 characters, contains "admin"
	if got := len([]rune(valor)); got != 20 {
		t.Fatalf("fixture is %d characters, want 20", got)
	}
	rep := PasswordReport(valor)
	if rep.Valido {
		t.Fatal("password with a forbidden word must be invalid")
	}
	if rep.Nivel == nil || *rep.Nivel != "Baja" {
		t.Errorf("nivel = %v, want Baja", rep.Nivel)
	}
	if p8 := ruleByID(rep, "P8"); p8 == nil || p8.OK || !p8.Evaluado {
		t.Errorf("P8 should fail and be evaluated: %+v", p8)
	}
}

// TestEmailLocalLengthBoundaries pins the RFC 5321 limit: 64 characters pass,
// 65 fail C3 only.
func TestEmailLocalLengthBoundaries(t *testing.T) {
	cases := []struct {
		n     int
		valid bool
	}{
		{64, true},
		{65, false},
	}
	for _, c := range cases {
		valor := strings.Repeat("a", c.n) + "@correo.com"
		rep := EmailReport(valor)
		if rep.Valido != c.valid {
			t.Errorf("parte local %d: valido = %v, want %v (%s)", c.n, rep.Valido, c.valid, failureSummary(rep))
		}
		if c.valid {
			continue
		}
		if r := ruleByID(rep, "C3"); r == nil || r.OK {
			t.Errorf("parte local %d: C3 should fail: %+v", c.n, r)
		}
		if r := ruleByID(rep, "C2"); r == nil || !r.OK {
			t.Errorf("parte local %d: C2 should still pass: %+v", c.n, r)
		}
	}
}

// TestEmailLabelBoundaries pins the 63-character label limit: 63 passes, 64
// fails C4.
func TestEmailLabelBoundaries(t *testing.T) {
	cases := []struct {
		n     int
		valid bool
	}{
		{63, true},
		{64, false},
	}
	for _, c := range cases {
		valor := "ana@" + strings.Repeat("a", c.n) + ".com"
		rep := EmailReport(valor)
		if rep.Valido != c.valid {
			t.Errorf("etiqueta %d: valido = %v, want %v (%s)", c.n, rep.Valido, c.valid, failureSummary(rep))
		}
		if !c.valid {
			if r := ruleByID(rep, "C4"); r == nil || r.OK {
				t.Errorf("etiqueta %d: C4 should fail: %+v", c.n, r)
			}
		}
	}
}

// TestEmailDomainLengthBoundaries pins the 255-octet domain limit: a domain of
// exactly 255 passes, 256 fails C6 while every label stays legal.
func TestEmailDomainLengthBoundaries(t *testing.T) {
	// 63 + 1 + 63 + 1 + 63 + 1 + 63 = 255, four legal labels.
	domain255 := strings.Join([]string{
		strings.Repeat("a", 63),
		strings.Repeat("b", 63),
		strings.Repeat("c", 63),
		strings.Repeat("d", 63),
	}, ".")
	if len(domain255) != 255 {
		t.Fatalf("fixture domain is %d characters, want 255", len(domain255))
	}

	// 63+1+63+1+63+1+50+1+13 = 256: one more octet, every label still <= 63.
	domain256 := strings.Join([]string{
		strings.Repeat("a", 63),
		strings.Repeat("b", 63),
		strings.Repeat("c", 63),
		strings.Repeat("d", 50),
		strings.Repeat("e", 13),
	}, ".")
	if len(domain256) != 256 {
		t.Fatalf("fixture domain is %d characters, want 256", len(domain256))
	}

	if rep := EmailReport("ana@" + domain255); !rep.Valido {
		t.Errorf("dominio de 255 debería ser válido (%s)", failureSummary(rep))
	}

	rep := EmailReport("ana@" + domain256)
	if rep.Valido {
		t.Error("dominio de 256 debería ser inválido")
	}
	if r := ruleByID(rep, "C6"); r == nil || r.OK {
		t.Errorf("C6 should fail on a 256-character domain: %+v", r)
	}
	if r := ruleByID(rep, "C4"); r == nil || !r.OK {
		t.Errorf("C4 should still pass: every label is <= 63 (%+v)", r)
	}
}

// TestEmailSpacesAreStripped mirrors the password behaviour: spaces are not part
// of the alphabet and are removed before validating.
func TestEmailSpacesAreStripped(t *testing.T) {
	rep := EmailReport("  ana.perez@correo.com  ")
	if !rep.Valido {
		t.Fatalf("should be valid after stripping spaces (%s)", failureSummary(rep))
	}
	if rep.Normalizado != "ana.perez@correo.com" {
		t.Errorf("normalizado = %q", rep.Normalizado)
	}
}

func failureSummary(rep Report) string {
	var parts []string
	for _, r := range rep.Reglas {
		if r.Evaluado && !r.OK {
			parts = append(parts, r.ID+": "+r.Mensaje)
		}
	}
	if len(parts) == 0 {
		return "sin fallas"
	}
	return strings.Join(parts, "; ")
}
