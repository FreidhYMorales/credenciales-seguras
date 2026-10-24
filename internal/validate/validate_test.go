package validate

import "testing"

func ruleByID(r Report, id string) *RuleResult {
	for i := range r.Reglas {
		if r.Reglas[i].ID == id {
			return &r.Reglas[i]
		}
	}
	return nil
}

func TestPasswordLevels(t *testing.T) {
	cases := []struct {
		valor string
		nivel string
		ok    bool
	}{
		{"Pl4t0n0#Guatemala2026", "Fuerte", true}, // 21 caracteres
		{"Miperrocome7Huesos!", "Media", true},    // 19 caracteres
		{"Xk9$mTq2#vLw8@Rz", "Media", true},       // 16 caracteres, ejemplo del anexo
		{"Abc123!", "Baja", false},                // incumple P1
	}
	for _, c := range cases {
		rep := PasswordReport(c.valor)
		if rep.Valido != c.ok {
			t.Errorf("%q: valido = %v, want %v", c.valor, rep.Valido, c.ok)
		}
		if rep.Nivel == nil || *rep.Nivel != c.nivel {
			got := "<nil>"
			if rep.Nivel != nil {
				got = *rep.Nivel
			}
			t.Errorf("%q: nivel = %s, want %s", c.valor, got, c.nivel)
		}
	}
}

func TestPasswordEntropy(t *testing.T) {
	rep := PasswordReport("Xk9$mTq2#vLw8@RzQe4&") // 20 caracteres
	if rep.Entropia == nil || *rep.Entropia != 131 {
		t.Fatalf("entropia = %v, want 131", rep.Entropia)
	}
	if *rep.Nivel != "Fuerte" {
		t.Errorf("nivel = %s, want Fuerte", *rep.Nivel)
	}
}

func TestPasswordReportAlwaysComplete(t *testing.T) {
	// Outside the alphabet: P0 fails and the other eight are reported as skipped.
	rep := PasswordReport("clave con ñ")
	if len(rep.Reglas) != 9 {
		t.Fatalf("%d reglas, want 9", len(rep.Reglas))
	}
	if p0 := ruleByID(rep, "P0"); p0 == nil || p0.OK || p0.Posicion == nil {
		t.Fatalf("P0 should fail with a position: %+v", p0)
	}
	for _, r := range rep.Reglas[1:] {
		if r.Evaluado {
			t.Errorf("%s should be skipped", r.ID)
		}
	}

	// Short password: P1 fails but the rest are still evaluated.
	rep = PasswordReport("Abc1!")
	if len(rep.Reglas) != 9 {
		t.Fatalf("%d reglas, want 9", len(rep.Reglas))
	}
	if p1 := ruleByID(rep, "P1"); p1 == nil || p1.OK || p1.Mensaje != "longitud = 5 (faltan 10)" {
		t.Errorf("P1 = %+v, want the faltan message", p1)
	}
	for _, r := range rep.Reglas[1:] {
		if !r.Evaluado {
			t.Errorf("%s should be evaluated", r.ID)
		}
	}
}

func TestPasswordRejectionFilters(t *testing.T) {
	cases := []struct {
		valor  string
		regla  string
		mensaj string
	}{
		{"Xk9$mTq2#vLw8@RzQaaa", "P6", "3 o más caracteres idénticos consecutivos"},
		{"Xk9$mTq2#vLw8@RzQabcd", "P7", "secuencia ascendente o descendente de 4"},
		{"Xk9$mTq2#vLw8@password", "P8", "contiene una palabra prohibida"},
	}
	for _, c := range cases {
		rep := PasswordReport(c.valor)
		r := ruleByID(rep, c.regla)
		if r == nil || r.OK || !r.Evaluado {
			t.Errorf("%q: %s should fail and be evaluated: %+v", c.valor, c.regla, r)
			continue
		}
		if r.Mensaje != c.mensaj {
			t.Errorf("%q: %s mensaje = %q, want %q", c.valor, c.regla, r.Mensaje, c.mensaj)
		}
		if rep.Valido {
			t.Errorf("%q: should be invalid", c.valor)
		}
		if rep.Nivel == nil || *rep.Nivel != "Baja" {
			t.Errorf("%q: nivel should be Baja", c.valor)
		}
	}
}

func TestSpacesAreStripped(t *testing.T) {
	rep := PasswordReport(" Xk9$mTq2# vLw8@RzQe4& ")
	if rep.Normalizado != "Xk9$mTq2#vLw8@RzQe4&" {
		t.Errorf("normalizado = %q", rep.Normalizado)
	}
	if rep.Longitud != 20 || !rep.Valido {
		t.Errorf("longitud = %d, valido = %v", rep.Longitud, rep.Valido)
	}
}

func TestEmailReport(t *testing.T) {
	cases := []struct {
		valor string
		ok    bool
		regla string // expected failing rule, "" when valid
	}{
		{"ana.perez@correo.com", true, ""},
		{"usuario+etiqueta@correo.com", true, ""},
		{"ana..perez@correo.com", false, "C2"},
		{".ana@correo.com", false, "C2"},
		{"ana@correo", false, "C4"},
		{"ana@-correo.com", false, "C4"},
		{"ana@correo.c", false, "C5"},
		{"ana@correo.c0m", false, "C5"},
		{"ana@correo.com@x", false, "C1"},
		{"ana@correo.com ", true, ""},
	}
	for _, c := range cases {
		rep := EmailReport(c.valor)
		if rep.Valido != c.ok {
			t.Errorf("%q: valido = %v, want %v", c.valor, rep.Valido, c.ok)
		}
		if len(rep.Reglas) != 7 {
			t.Errorf("%q: %d reglas, want 7", c.valor, len(rep.Reglas))
		}
		if rep.Nivel != nil || rep.Entropia != nil {
			t.Errorf("%q: level and entropy must be null for emails", c.valor)
		}
		if c.regla != "" {
			r := ruleByID(rep, c.regla)
			if r == nil || r.OK {
				t.Errorf("%q: %s should fail: %+v", c.valor, c.regla, r)
			}
		}
	}
}

func TestEmailShortCircuitKeepsAllRules(t *testing.T) {
	rep := EmailReport("sin-arroba")
	if len(rep.Reglas) != 7 {
		t.Fatalf("%d reglas, want 7", len(rep.Reglas))
	}
	if c1 := ruleByID(rep, "C1"); c1 == nil || c1.OK {
		t.Fatalf("C1 should fail: %+v", c1)
	}
	for _, r := range rep.Reglas[2:] {
		if r.Evaluado {
			t.Errorf("%s should be skipped after a C1 failure", r.ID)
		}
	}
}

func TestEmailAlphabet(t *testing.T) {
	rep := EmailReport("ana@correo.com;")
	if p0 := ruleByID(rep, "C0"); p0 == nil || p0.OK || p0.Posicion == nil {
		t.Fatalf("C0 should fail on ';' with a position: %+v", p0)
	} else if *p0.Posicion != 14 {
		t.Errorf("position = %d, want 14", *p0.Posicion)
	}
}
