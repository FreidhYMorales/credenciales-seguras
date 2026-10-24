package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FreidhYMorales/credenciales-seguras/internal/validate"
	"github.com/FreidhYMorales/credenciales-seguras/testdata"
)

// TestPairsRegenerate is the byte-for-byte gate: rebuilding the 20 pairs from
// their recorded seeds must reproduce testdata/pares.json exactly. If a change
// to the engine, the rules or the generator alters a graded value, this fails
// before the file is silently rewritten.
func TestPairsRegenerate(t *testing.T) {
	stored, err := testdata.FS.ReadFile("pares.json")
	if err != nil {
		t.Fatalf("embedded pares.json: %v", err)
	}

	pairs, err := testdata.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	regenerated, err := testdata.Encode(pairs)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	if !bytes.Equal(stored, regenerated) {
		t.Fatalf("pares.json does not regenerate byte for byte (stored %d bytes, rebuilt %d bytes)",
			len(stored), len(regenerated))
	}
}

// TestPairsVerdicts checks the 20 pairs against the shipped validator: every
// verdict matches and every invalid pair names the rule that actually fails.
func TestPairsVerdicts(t *testing.T) {
	pairs, err := testdata.Pairs()
	if err != nil {
		t.Fatalf("pairs: %v", err)
	}
	if len(pairs) != 20 {
		t.Fatalf("%d pairs, want 20", len(pairs))
	}

	valid := 0
	for _, p := range pairs {
		er := validate.EmailReport(p.Email)
		pr := validate.PasswordReport(p.Password)
		if er.Valido != p.EmailEsperado {
			t.Errorf("par %d: correo %q válido=%v, esperado=%v", p.ID, p.Email, er.Valido, p.EmailEsperado)
		}
		if pr.Valido != p.PasswordEsperado {
			t.Errorf("par %d: contraseña %q válida=%v, esperado=%v", p.ID, p.Password, pr.Valido, p.PasswordEsperado)
		}
		if p.EmailEsperado && p.PasswordEsperado {
			valid++
			if p.SeedEmail == nil || p.SeedPassword == nil {
				t.Errorf("par %d: un par válido debe registrar sus dos semillas", p.ID)
			}
			if p.ReglaFallida != nil {
				t.Errorf("par %d: un par válido no tiene regla fallida", p.ID)
			}
			continue
		}
		if p.ReglaFallida == nil {
			t.Errorf("par %d: un par inválido debe nombrar la regla que falla", p.ID)
			continue
		}
		if p.SeedEmail != nil || p.SeedPassword != nil {
			t.Errorf("par %d: un par redactado a mano no lleva semillas", p.ID)
		}
		if !failsRule(pr.Reglas, *p.ReglaFallida) && !failsRule(er.Reglas, *p.ReglaFallida) {
			t.Errorf("par %d: la regla %s no falla", p.ID, *p.ReglaFallida)
		}
	}
	if valid != 10 {
		t.Errorf("%d valid pairs, want 10", valid)
	}
}

func failsRule(rs []validate.RuleResult, id string) bool {
	for _, r := range rs {
		if r.ID == id && r.Evaluado && !r.OK {
			return true
		}
	}
	return false
}

// TestPairsEndpoint checks the HTTP surface of GET /api/pares: the envelope,
// the 20 entries and the field names the frontend reads.
func TestPairsEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pares", nil)
	New().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got struct {
		OK    bool `json:"ok"`
		Datos []struct {
			ID               int     `json:"id"`
			Email            string  `json:"email"`
			Password         string  `json:"password"`
			EmailEsperado    bool    `json:"email_esperado"`
			PasswordEsperado bool    `json:"password_esperado"`
			ReglaFallida     *string `json:"regla_fallida"`
			SeedEmail        *int64  `json:"seed_email"`
			SeedPassword     *int64  `json:"seed_password"`
		} `json:"datos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !got.OK {
		t.Fatal("ok = false, want true")
	}
	if len(got.Datos) != 20 {
		t.Fatalf("%d entries, want 20", len(got.Datos))
	}
	for i, p := range got.Datos {
		if p.ID != i+1 {
			t.Errorf("entry %d has id %d", i, p.ID)
		}
		if p.Email == "" || p.Password == "" {
			t.Errorf("par %d: empty credential", p.ID)
		}
	}
}

// TestValidarEndpointRejectsBadInput covers the 4xx side of the contract: an
// invalid request is an error, not a validation verdict.
func TestValidarEndpointRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"unknown tipo", `{"tipo":"otra","valor":"x"}`, http.StatusBadRequest},
		{"empty valor", `{"tipo":"password","valor":""}`, http.StatusBadRequest},
		{"broken json", `{"tipo":`, http.StatusBadRequest},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/validar", bytes.NewBufferString(c.body))
		New().ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s: status = %d, want %d (body %s)", c.name, rec.Code, c.want, rec.Body.String())
		}
		var env struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Errorf("%s: invalid JSON: %v", c.name, err)
			continue
		}
		if env.OK || env.Error == "" {
			t.Errorf("%s: envelope = %+v, want ok=false with an error message", c.name, env)
		}
	}
}

// TestValidarEndpointOutcomeIs200 pins the other half: a failed validation is
// an answer, so it must be HTTP 200.
func TestValidarEndpointOutcomeIs200(t *testing.T) {
	body := `{"tipo":"password","valor":"Abc123!"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/validar", bytes.NewBufferString(body))
	New().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var env struct {
		OK    bool            `json:"ok"`
		Datos validate.Report `json:"datos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !env.OK {
		t.Fatal("ok = false, want true")
	}
	if env.Datos.Valido {
		t.Error("valido = true, want false")
	}
	if len(env.Datos.Reglas) != 9 {
		t.Errorf("%d rules, want 9", len(env.Datos.Reglas))
	}
}

// TestGenerarEndpointDeterministic: the same seed through HTTP returns the same
// value, which is what the demo and the video replay.
func TestGenerarEndpointDeterministic(t *testing.T) {
	gen := func(body string) string {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/generar", bytes.NewBufferString(body))
		New().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
		}
		var env struct {
			OK    bool `json:"ok"`
			Datos struct {
				Valor   string `json:"valor"`
				Semilla int64  `json:"semilla"`
			} `json:"datos"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if !env.OK || env.Datos.Valor == "" {
			t.Fatalf("generation failed: %s", rec.Body.String())
		}
		if env.Datos.Semilla != 2026 {
			t.Errorf("semilla = %d, want 2026", env.Datos.Semilla)
		}
		return env.Datos.Valor
	}

	a := gen(`{"tipo":"password","longitud":20,"semilla":2026}`)
	b := gen(`{"tipo":"password","longitud":20,"semilla":2026}`)
	if a != b {
		t.Fatalf("same seed returned %q and %q", a, b)
	}
	if a != ">nslx613LK2%|#PxFbA5" {
		t.Errorf("demo value = %q, want the recorded one", a)
	}
}

// TestGenerarEndpointRejectsRange pins the documented 15-64 range.
func TestGenerarEndpointRejectsRange(t *testing.T) {
	for _, body := range []string{
		`{"tipo":"password","longitud":14,"semilla":1}`,
		`{"tipo":"password","longitud":65,"semilla":1}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/generar", bytes.NewBufferString(body))
		New().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", body, rec.Code)
		}
	}
}
