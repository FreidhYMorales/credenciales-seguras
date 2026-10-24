package api

import (
	"bytes"
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

// TestSubgrafo pins the BFS neighborhood served for views too large to draw.
// P1/min is the flagship case: 65 states, no pre-generated SVG, so the browser
// can only draw it from this payload.
func TestSubgrafo(t *testing.T) {
	type nodo struct {
		ID        string `json:"id"`
		Inicial   bool   `json:"inicial"`
		Aceptor   bool   `json:"aceptor"`
		Distancia int    `json:"distancia"`
	}
	type respuesta struct {
		Ok    bool `json:"ok"`
		Datos struct {
			Subgrafo *struct {
				Centro      string `json:"centro"`
				Profundidad int    `json:"profundidad"`
				Nodos       []nodo `json:"nodos"`
				Truncado    bool   `json:"truncado"`
				TotalNodos  int    `json:"total_nodos"`
			} `json:"subgrafo"`
		} `json:"datos"`
	}

	pedir := func(t *testing.T, ruta string) (int, []byte) {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, ruta, nil)
		New().ServeHTTP(rec, req)
		return rec.Code, rec.Body.Bytes()
	}

	const ruta = "/api/automatas/P1?vista=min&profundidad=1"
	codigo, cuerpo := pedir(t, ruta)
	if codigo != http.StatusOK {
		t.Fatalf("status = %d, want %d (body %s)", codigo, http.StatusOK, cuerpo)
	}

	var got respuesta
	if err := json.Unmarshal(cuerpo, &got); err != nil {
		t.Fatalf("invalid JSON: %v (body %s)", err, cuerpo)
	}
	sg := got.Datos.Subgrafo
	if sg == nil {
		t.Fatal("subgrafo = null, want an object")
	}
	if sg.Centro != "q0" {
		t.Errorf("centro = %q, want %q", sg.Centro, "q0")
	}
	if sg.Profundidad != 1 {
		t.Errorf("profundidad = %d, want 1", sg.Profundidad)
	}
	if len(sg.Nodos) == 0 {
		t.Fatal("subgrafo.nodos está vacío")
	}
	if p := sg.Nodos[0]; p.ID != "q0" || !p.Inicial || p.Distancia != 0 {
		t.Errorf("primer nodo = %+v, want q0 inicial a distancia 0", p)
	}
	for _, n := range sg.Nodos {
		if n.Distancia > 1 {
			t.Errorf("%s: distancia = %d, want <= 1", n.ID, n.Distancia)
		}
	}
	if sg.TotalNodos != 65 {
		t.Errorf("total_nodos = %d, want 65", sg.TotalNodos)
	}

	// Determinism: the same request twice must return the same bytes.
	if _, otro := pedir(t, ruta); !bytes.Equal(cuerpo, otro) {
		t.Errorf("dos pedidos idénticos devolvieron cuerpos distintos")
	}

	// An unknown center is a malformed request.
	if codigo, _ := pedir(t, "/api/automatas/P1?vista=min&profundidad=1&centro=q9999"); codigo != http.StatusBadRequest {
		t.Errorf("centro desconocido: status = %d, want %d", codigo, http.StatusBadRequest)
	}

	// Without profundidad there is no neighborhood at all.
	_, cuerpo = pedir(t, "/api/automatas/P1?vista=min")
	var sinProfundidad struct {
		Datos map[string]json.RawMessage `json:"datos"`
	}
	if err := json.Unmarshal(cuerpo, &sinProfundidad); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if raw := string(sinProfundidad.Datos["subgrafo"]); raw != "null" {
		t.Errorf("subgrafo = %s, want null", raw)
	}
}

// TestSubgrafoAncho covers the wide case: P6 has 59 outgoing edges from q0, so
// the ball must cap each ring, keep growing from the nodes that made it, and
// fall back to the unfolding instead of sending thousands of edges.
func TestSubgrafoAncho(t *testing.T) {
	type respuesta struct {
		Datos struct {
			Subgrafo struct {
				Nodos []struct {
					Distancia int `json:"distancia"`
				} `json:"nodos"`
				Aristas        []any `json:"aristas"`
				AristasTotales int   `json:"aristas_totales"`
				Truncado       bool  `json:"truncado"`
			} `json:"subgrafo"`
		} `json:"datos"`
	}

	pedir := func(t *testing.T, ruta string) respuesta {
		t.Helper()
		rec := httptest.NewRecorder()
		New().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d", ruta, rec.Code)
		}
		var got respuesta
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		return got
	}

	uno := pedir(t, "/api/automatas/P6?vista=min&profundidad=1")
	tres := pedir(t, "/api/automatas/P6?vista=min&profundidad=3")

	if !uno.Datos.Subgrafo.Truncado {
		t.Error("profundidad 1: truncado = false, want true (q0 has more children than the cap)")
	}
	if len(tres.Datos.Subgrafo.Nodos) <= len(uno.Datos.Subgrafo.Nodos) {
		t.Errorf("profundidad 3 devolvió %d nodos, want más que los %d de profundidad 1",
			len(tres.Datos.Subgrafo.Nodos), len(uno.Datos.Subgrafo.Nodos))
	}
	maximo := 0
	for _, n := range tres.Datos.Subgrafo.Nodos {
		if n.Distancia > maximo {
			maximo = n.Distancia
		}
	}
	if maximo != 3 {
		t.Errorf("profundidad 3 llegó a distancia %d, want 3", maximo)
	}
	if len(tres.Datos.Subgrafo.Aristas) >= tres.Datos.Subgrafo.AristasTotales {
		t.Errorf("aristas = %d, aristas_totales = %d; want the unfolding to hide some",
			len(tres.Datos.Subgrafo.Aristas), tres.Datos.Subgrafo.AristasTotales)
	}
}
