// Package api exposes the HTTP surface of the application.
//
// Every response uses the envelope {"ok": true, "datos": ...} or
// {"ok": false, "error": "..."}. Validation outcomes are HTTP 200; 4xx and
// 5xx are reserved for real errors.
package api

import (
	"encoding/json"
	"net/http"
)

// Server holds the dependencies shared by the handlers.
type Server struct{}

// New returns the configured HTTP handler.
func New() http.Handler {
	s := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/salud", s.salud)
	return mux
}

// salud is the liveness probe used by the smoke test and the run sheet.
func (s *Server) salud(w http.ResponseWriter, r *http.Request) {
	writeData(w, map[string]string{"estado": "ok"})
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
