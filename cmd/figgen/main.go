// Command figgen renders the automata figures to SVG with Graphviz, at build
// time, so the demo machine never needs Graphviz.
//
// Only the views that pass the drawing threshold are rendered: at most 40 states
// and 80 edges. Above that the explorer shows counts plus the transition table.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
)

const (
	maxStates = 40
	maxEdges  = 80
	outDir    = "assets/figures"
)

func main() {
	if _, err := exec.LookPath("dot"); err != nil {
		log.Fatal("graphviz 'dot' not found in PATH")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	rendered, skipped := 0, 0
	for _, r := range rules.All() {
		if !r.TieneER {
			continue
		}
		views := []struct {
			name  string
			dot   string
			nodes int
			edges int
		}{
			{"nfa", automata.DOTNFA(r.NFA), r.NFA.N, len(r.NFA.Trans)},
			{"sub", automata.DOTDFA(r.Subset), r.Subset.N, r.Subset.Edges()},
			{"tree", automata.DOTDFA(r.Tree), r.Tree.N, r.Tree.Edges()},
			{"min", automata.DOTDFA(r.Min), r.Min.N, r.Min.Edges()},
		}
		for _, v := range views {
			name := fmt.Sprintf("%s_%s.svg", safeID(r.ID), v.name)
			path := filepath.Join(outDir, name)
			if v.nodes > maxStates || v.edges > maxEdges {
				skipped++
				// Remove a stale figure from an earlier build.
				_ = os.Remove(path)
				continue
			}
			if err := render(v.dot, path); err != nil {
				log.Fatalf("%s/%s: %v", r.ID, v.name, err)
			}
			rendered++
		}
	}
	log.Printf("figures: %d rendered, %d skipped (threshold %d states / %d edges)", rendered, skipped, maxStates, maxEdges)
}

func render(dotSrc, path string) error {
	cmd := exec.Command("dot", "-Tsvg", "-o", path)
	cmd.Stdin = strings.NewReader(dotSrc)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func safeID(id string) string {
	return strings.ReplaceAll(id, "*", "star")
}
