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
