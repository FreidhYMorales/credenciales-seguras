// Package api exposes the HTTP surface of the application.
//
// Every response uses the envelope {"ok": true, "datos": ...} or
// {"ok": false, "error": "..."}. Validation outcomes are HTTP 200; 4xx and
// 5xx are reserved for real errors.
package api

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/fs"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/FreidhYMorales/credenciales-seguras/assets"
	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
	"github.com/FreidhYMorales/credenciales-seguras/internal/generate"
	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
	"github.com/FreidhYMorales/credenciales-seguras/internal/validate"
	"github.com/FreidhYMorales/credenciales-seguras/testdata"
	"github.com/FreidhYMorales/credenciales-seguras/web"
)

// Server holds the dependencies shared by the handlers.
type Server struct {
	web fs.FS
}

// New returns the configured HTTP handler.
func New() http.Handler {
	s := &Server{web: web.FS}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/salud", s.salud)
	mux.HandleFunc("POST /api/validar", s.validar)
	mux.HandleFunc("POST /api/generar", s.generar)
	mux.HandleFunc("GET /api/reglas", s.reglas)
	mux.HandleFunc("GET /api/automatas/{id}", s.automata)
	mux.HandleFunc("POST /api/simular", s.simular)
	mux.HandleFunc("GET /api/pares", s.pares)
	mux.HandleFunc("GET /", s.static)
	return mux
}

func (s *Server) salud(w http.ResponseWriter, r *http.Request) {
	writeData(w, map[string]string{"estado": "ok"})
}

func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		name = "index.html"
	}
	data, err := fs.ReadFile(s.web, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType(name))
	_, _ = w.Write(data)
}

func contentType(name string) string {
	switch {
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// ---------------------------------------------------------------- validar

type validarRequest struct {
	Tipo  string `json:"tipo"`
	Valor string `json:"valor"`
}

func (s *Server) validar(w http.ResponseWriter, r *http.Request) {
	var req validarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	kind := validate.Kind(req.Tipo)
	if kind != validate.Password && kind != validate.Email {
		writeError(w, http.StatusBadRequest, `tipo debe ser "password" o "email"`)
		return
	}
	if req.Valor == "" {
		writeError(w, http.StatusBadRequest, "valor vacío")
		return
	}
	writeData(w, validate.Validate(kind, req.Valor))
}

// ---------------------------------------------------------------- generar

type generarRequest struct {
	Tipo     string `json:"tipo"`
	Longitud *int   `json:"longitud"`
	Semilla  *int64 `json:"semilla"`
}

func (s *Server) generar(w http.ResponseWriter, r *http.Request) {
	var req generarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	kind := validate.Kind(req.Tipo)
	if kind != validate.Password && kind != validate.Email {
		writeError(w, http.StatusBadRequest, `tipo debe ser "password" o "email"`)
		return
	}

	seed := int64(0)
	if req.Semilla != nil {
		seed = *req.Semilla
	} else {
		var err error
		if seed, err = randomSeed(); err != nil {
			writeError(w, http.StatusInternalServerError, "no se pudo sortear la semilla")
			return
		}
	}

	var (
		res generate.Result
		err error
	)
	if kind == validate.Password {
		length := 20
		if req.Longitud != nil {
			length = *req.Longitud
		}
		res, err = generate.Password(seed, length)
	} else {
		res, err = generate.Email(seed)
	}
	if err != nil {
		if errors.Is(err, generate.ErrExhausted) {
			writeError(w, http.StatusInternalServerError, "no se pudo generar una credencial válida")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeData(w, res)
}

// ---------------------------------------------------------------- reglas

type reglaPayload struct {
	ID      string          `json:"id"`
	Nombre  string          `json:"nombre"`
	Tipo    string          `json:"tipo"`
	TieneER bool            `json:"tiene_er"`
	ER      *string         `json:"er"`
	RE2     *string         `json:"re2"`
	Conteos *automata.Stats `json:"conteos"`
}

func (s *Server) reglas(w http.ResponseWriter, r *http.Request) {
	all := rules.All()
	out := make([]reglaPayload, 0, len(all))
	for _, r := range all {
		p := reglaPayload{ID: r.ID, Nombre: r.Nombre, Tipo: r.Tipo, TieneER: r.TieneER}
		if r.TieneER {
			er, re2 := r.ER, r.RE2()
			counts := r.Counts()
			p.ER, p.RE2, p.Conteos = &er, &re2, &counts
		}
		out = append(out, p)
	}
	writeData(w, out)
}

// ------------------------------------------------------------- automatas

type transicion struct {
	Desde   string `json:"desde"`
	Simbolo string `json:"simbolo"`
	Hacia   string `json:"hacia"`
}

func (s *Server) automata(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rule := rules.Get(id)
	if rule == nil {
		writeError(w, http.StatusNotFound, "regla desconocida: "+id)
		return
	}
	if !rule.TieneER {
		writeError(w, http.StatusBadRequest, id+" es una verificación de conjunto y no tiene autómata")
		return
	}

	vista := r.URL.Query().Get("vista")
	if vista == "" {
		vista = "sub"
	}
	pagina := intParam(r, "pagina", 1)
	if pagina < 1 {
		pagina = 1
	}
	limite := intParam(r, "limite", 50)
	if limite < 1 || limite > 200 {
		limite = 50
	}

	transiciones, total, err := transitions(rule, vista)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// profundidad is optional: absent or 0 means "no neighborhood". Any other
	// value is clamped to the supported 1..3 range.
	profundidad := intParam(r, "profundidad", 0)
	if profundidad != 0 {
		profundidad = max(1, min(profundidad, 3))
	}
	var sg *subgrafo
	if profundidad > 0 {
		centro := r.URL.Query().Get("centro")
		if centro == "" {
			centro = "q0"
		}
		if sg, err = neighborhood(rule, vista, centro, profundidad); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	start := (pagina - 1) * limite
	if start > len(transiciones) {
		start = len(transiciones)
	}
	end := start + limite
	if end > len(transiciones) {
		end = len(transiciones)
	}

	payload := map[string]any{
		"id":           id,
		"vista":        vista,
		"svg":          svgFor(id, vista),
		"motivo":       motivo(rule, vista),
		"subgrafo":     sg,
		"transiciones": transiciones[start:end],
		"total":        total,
		"pagina":       pagina,
		"limite":       limite,
	}
	writeData(w, payload)
}

func transitions(rule *rules.Rule, vista string) ([]transicion, int, error) {
	if vista == "nfa" {
		out := make([]transicion, 0, len(rule.NFA.Trans))
		for _, t := range rule.NFA.Trans {
			label := t.Name
			if t.Chars == nil {
				label = "ε"
			}
			out = append(out, transicion{
				Desde:   "q" + strconv.Itoa(t.From),
				Simbolo: label,
				Hacia:   "q" + strconv.Itoa(t.To),
			})
		}
		return out, len(out), nil
	}
	d, err := viewDFA(rule, vista)
	if err != nil {
		return nil, 0, err
	}
	out := make([]transicion, 0, d.N*len(d.Atoms))
	for q := 0; q < d.N; q++ {
		for a := range d.Atoms {
			t := d.Delta[q][a]
			if t < 0 {
				continue
			}
			out = append(out, transicion{
				Desde:   "q" + strconv.Itoa(q),
				Simbolo: automata.AtomLabel(d.Atoms[a]),
				Hacia:   "q" + strconv.Itoa(t),
			})
		}
	}
	return out, len(out), nil
}

func viewDFA(rule *rules.Rule, vista string) (*automata.DFA, error) {
	switch vista {
	case "sub":
		return rule.Subset, nil
	case "tree":
		return rule.Tree, nil
	case "min":
		return rule.Min, nil
	default:
		return nil, errors.New(`vista debe ser "nfa", "sub", "tree" o "min"`)
	}
}

func svgFor(id, vista string) *string {
	name := strings.ReplaceAll(id, "*", "star") + "_" + vista + ".svg"
	raw, err := assets.FS.ReadFile(assets.FiguresDir + "/" + name)
	if err != nil {
		return nil
	}
	s := string(raw)
	return &s
}

func motivo(rule *rules.Rule, vista string) *string {
	if svgFor(rule.ID, vista) != nil {
		return nil
	}
	n, e := viewSize(rule, vista)
	if n == 0 {
		return nil
	}
	msg := "la vista tiene " + strconv.Itoa(n) + " estados y " + strconv.Itoa(e) + " transiciones; se muestra la tabla"
	return &msg
}

func viewSize(rule *rules.Rule, vista string) (int, int) {
	if vista == "nfa" {
		return rule.NFA.N, len(rule.NFA.Trans)
	}
	d, err := viewDFA(rule, vista)
	if err != nil {
		return 0, 0
	}
	return d.N, d.Edges()
}

// ------------------------------------------------------------ subgrafo

// The neighborhood is drawn as a breadth-first unfolding: one parent edge per
// node. Drawing every edge of the induced subgraph is unreadable for a wide
// automaton (P6 has 3481 edges among its first 60 states), so a ball above
// maxAristasDibujables falls back to the tree and only reports the count.
const (
	maxSubgrafoNodos     = 60 // hard cap on the whole ball
	maxSubgrafoPorNivel  = 10 // keeps one row inside the card without shrinking it
	maxAristasDibujables = 60 // above this, the ball is drawn as a tree
)

type subgrafoNodo struct {
	ID        string `json:"id"`
	Inicial   bool   `json:"inicial"`
	Aceptor   bool   `json:"aceptor"`
	Distancia int    `json:"distancia"`
}

type subgrafoArista struct {
	Desde   string `json:"desde"`
	Simbolo string `json:"simbolo"`
	Hacia   string `json:"hacia"`
}

type subgrafo struct {
	Centro         string           `json:"centro"`
	Profundidad    int              `json:"profundidad"`
	Nodos          []subgrafoNodo   `json:"nodos"`
	Aristas        []subgrafoArista `json:"aristas"`         // parent edge of every node but the centre
	AristasTotales int              `json:"aristas_totales"` // edges of the induced subgraph
	Truncado       bool             `json:"truncado"`
	TotalNodos     int              `json:"total_nodos"`
}

// vecino is one outgoing edge of a view state.
type vecino struct {
	origen  int
	simbolo string
	hacia   int
}

// vistaGrafo is the uniform view of one automaton that the BFS walks, so the
// NFA and the three DFA views share the same neighborhood code.
type vistaGrafo struct {
	n       int
	inicial int
	aceptor []bool
	salidas [][]vecino
}

func grafoDe(rule *rules.Rule, vista string) (*vistaGrafo, error) {
	if vista == "nfa" {
		nfa := rule.NFA
		g := &vistaGrafo{
			n:       nfa.N,
			inicial: nfa.Start,
			aceptor: make([]bool, nfa.N),
			salidas: make([][]vecino, nfa.N),
		}
		g.aceptor[nfa.Accept] = true
		for _, t := range nfa.Trans {
			label := t.Name
			if t.Chars == nil {
				label = "ε"
			}
			g.salidas[t.From] = append(g.salidas[t.From], vecino{origen: t.From, simbolo: label, hacia: t.To})
		}
		for _, vs := range g.salidas {
			sortVecinos(vs)
		}
		return g, nil
	}

	d, err := viewDFA(rule, vista)
	if err != nil {
		return nil, err
	}
	g := &vistaGrafo{
		n:       d.N,
		inicial: d.Start,
		aceptor: d.Accept,
		salidas: make([][]vecino, d.N),
	}
	for q := 0; q < d.N; q++ {
		for a := range d.Atoms {
			t := d.Delta[q][a]
			if t < 0 {
				continue
			}
			g.salidas[q] = append(g.salidas[q], vecino{origen: q, simbolo: automata.AtomLabel(d.Atoms[a]), hacia: t})
		}
		sortVecinos(g.salidas[q])
	}
	return g, nil
}

// sortVecinos orders a state's edges by (symbol, target) so two identical
// requests visit the frontier in the same order.
func sortVecinos(vs []vecino) {
	sort.SliceStable(vs, func(i, j int) bool {
		if vs[i].simbolo != vs[j].simbolo {
			return vs[i].simbolo < vs[j].simbolo
		}
		if vs[i].hacia != vs[j].hacia {
			return vs[i].hacia < vs[j].hacia
		}
		return vs[i].origen < vs[j].origen
	})
}

// estadoID parses a state id of the form "q<number>".
func estadoID(s string) (int, bool) {
	if !strings.HasPrefix(s, "q") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(s, "q"))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// neighborhood returns the BFS ball of radius profundidad around centro.
func neighborhood(rule *rules.Rule, vista, centro string, profundidad int) (*subgrafo, error) {
	g, err := grafoDe(rule, vista)
	if err != nil {
		return nil, err
	}
	inicio, ok := estadoID(centro)
	if !ok || inicio >= g.n {
		return nil, errors.New("centro desconocido: " + centro)
	}

	visto := make([]bool, g.n)
	visto[inicio] = true
	sg := &subgrafo{
		Centro:      "q" + strconv.Itoa(inicio),
		Profundidad: profundidad,
		TotalNodos:  g.n,
		Nodos: []subgrafoNodo{{
			ID:        "q" + strconv.Itoa(inicio),
			Inicial:   g.inicial == inicio,
			Aceptor:   g.aceptor[inicio],
			Distancia: 0,
		}},
	}

	// padre[q] is the edge that discovered q, so the drawing has exactly one
	// incoming edge per node and stays legible.
	padre := make(map[int]subgrafoArista, maxSubgrafoNodos)

	frontera := []int{inicio}
	for d := 1; d <= profundidad && len(frontera) > 0; d++ {
		candidatos := make([]vecino, 0, len(frontera))
		for _, q := range frontera {
			candidatos = append(candidatos, g.salidas[q]...)
		}
		sortVecinos(candidatos)

		siguiente := make([]int, 0, len(candidatos))
		nivel := 0
		lleno := false
		for _, v := range candidatos {
			if visto[v.hacia] {
				continue
			}
			if nivel >= maxSubgrafoPorNivel || len(sg.Nodos) >= maxSubgrafoNodos {
				sg.Truncado = true
				lleno = len(sg.Nodos) >= maxSubgrafoNodos
				break
			}
			visto[v.hacia] = true
			nivel++
			siguiente = append(siguiente, v.hacia)
			sg.Nodos = append(sg.Nodos, subgrafoNodo{
				ID:        "q" + strconv.Itoa(v.hacia),
				Inicial:   g.inicial == v.hacia,
				Aceptor:   g.aceptor[v.hacia],
				Distancia: d,
			})
			padre[v.hacia] = subgrafoArista{
				Desde:   "q" + strconv.Itoa(v.origen),
				Simbolo: v.simbolo,
				Hacia:   "q" + strconv.Itoa(v.hacia),
			}
		}
		if lleno {
			break // no room for another ring
		}
		// A capped level still lets the next one grow from the nodes that made
		// it, so raising the depth keeps revealing structure.
		frontera = siguiente
	}

	sg.Aristas = make([]subgrafoArista, 0, len(sg.Nodos)-1)
	todas := make([]subgrafoArista, 0, len(sg.Nodos))
	for _, n := range sg.Nodos {
		q, _ := estadoID(n.ID)
		if a, ok := padre[q]; ok {
			sg.Aristas = append(sg.Aristas, a)
		}
		for _, v := range g.salidas[q] {
			if !visto[v.hacia] {
				continue
			}
			todas = append(todas, subgrafoArista{
				Desde:   n.ID,
				Simbolo: v.simbolo,
				Hacia:   "q" + strconv.Itoa(v.hacia),
			})
		}
	}
	sg.AristasTotales = len(todas)
	// A small ball is clearer with every edge; a wide one is a hairball, so it
	// falls back to the unfolding. The count is always reported.
	if len(todas) <= maxAristasDibujables {
		sg.Aristas = todas
	}
	return sg, nil
}

// -------------------------------------------------------------- simular

type simularRequest struct {
	ID    string `json:"id"`
	Valor string `json:"valor"`
	Vista string `json:"vista"`
}

type trazaPaso struct {
	Paso    int    `json:"paso"`
	Estado  string `json:"estado"`
	Simbolo string `json:"simbolo"`
}

func (s *Server) simular(w http.ResponseWriter, r *http.Request) {
	var req simularRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return
	}
	rule := rules.Get(req.ID)
	if rule == nil || !rule.TieneER {
		writeError(w, http.StatusNotFound, "regla desconocida: "+req.ID)
		return
	}
	if req.Vista == "" {
		req.Vista = "sub"
	}
	d, err := viewDFA(rule, req.Vista)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	acepta, pasos, fallo := automata.Simulate(d, []rune(req.Valor))
	traza := make([]trazaPaso, 0, len(pasos))
	for _, p := range pasos {
		traza = append(traza, trazaPaso{Paso: p.Paso, Estado: "q" + strconv.Itoa(p.Desde), Simbolo: string(p.Simbolo)})
	}
	var falloPtr *int
	if fallo >= 0 {
		falloPtr = &fallo
	}
	writeData(w, map[string]any{"acepta": acepta, "traza": traza, "fallo": falloPtr})
}

// ---------------------------------------------------------------- pares

func (s *Server) pares(w http.ResponseWriter, r *http.Request) {
	pairs, err := testdata.Pairs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo leer pares.json")
		return
	}
	writeData(w, pairs)
}

// ---------------------------------------------------------------- utils

func intParam(r *http.Request, name string, def int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// randomSeed draws a non-negative seed when the client does not supply one, so
// the generation can still be replayed afterwards.
func randomSeed() (int64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b[:]) & math.MaxInt64), nil
}

func writeData(w http.ResponseWriter, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "datos": datos})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": msg})
}
