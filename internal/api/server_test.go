package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSalud is the bootstrap smoke test: it proves the server builds, routes
// and honours the response envelope.
func TestSalud(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/salud", nil)

	New().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got struct {
		Ok    bool              `json:"ok"`
		Datos map[string]string `json:"datos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v (body %q)", err, rec.Body.String())
	}
	if !got.Ok {
		t.Errorf("ok = false, want true")
	}
	if got.Datos["estado"] != "ok" {
		t.Errorf("datos.estado = %q, want %q", got.Datos["estado"], "ok")
	}
}

// TestReglasConteosKeys pins the JSON shape of the state counts to the one
// documented in docs/api.md. The struct carries no omitempty, so a missing tag
// would surface as "NFA"/"Subset"/"Tree"/"Min" and the frontend would render
// "undefined" next to every rule.
func TestReglasConteosKeys(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/reglas", nil)

	New().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got struct {
		Datos []struct {
			ID      string                     `json:"id"`
			Conteos map[string]json.RawMessage `json:"conteos"`
		} `json:"datos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	quiero := []string{"nfa", "sub", "tree", "min"}
	for _, r := range got.Datos {
		if r.Conteos == nil {
			continue // P0 and C0 have no automaton.
		}
		if len(r.Conteos) != len(quiero) {
			t.Errorf("%s: %d claves en conteos, want %d", r.ID, len(r.Conteos), len(quiero))
		}
		for _, k := range quiero {
			if _, ok := r.Conteos[k]; !ok {
				t.Errorf("%s: falta la clave %q en conteos", r.ID, k)
			}
		}
	}

	// P1 is the largest automaton and the one the document quotes.
	encontrado := false
	for _, r := range got.Datos {
		if r.ID != "P1" {
			continue
		}
		encontrado = true
		var c struct {
			NFA  int `json:"nfa"`
			Sub  int `json:"sub"`
			Tree int `json:"tree"`
			Min  int `json:"min"`
		}
		raw, _ := json.Marshal(r.Conteos)
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("conteos de P1: %v", err)
		}
		if c.NFA != 1092 || c.Sub != 257 || c.Tree != 65 || c.Min != 65 {
			t.Errorf("conteos de P1 = %d/%d/%d/%d, want 1092/257/65/65", c.NFA, c.Sub, c.Tree, c.Min)
		}
	}
	if !encontrado {
		t.Error("la regla P1 no aparece en /api/reglas")
	}
}
