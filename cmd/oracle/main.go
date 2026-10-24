// Command oracle answers a corpus of (rule, string) cases with the Go engine and
// prints the state counts of every rule. It is the Go half of the differential
// test against the Python prototype (tools/oracle/check.py); it is never part of
// the running application.
//
// Protocol: a JSON object {"casos":[{"id":"P1","valor":"..."}]} on stdin, a JSON
// object {"conteos":{...},"veredictos":[...]} on stdout.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FreidhYMorales/credenciales-seguras/internal/automata"
	"github.com/FreidhYMorales/credenciales-seguras/internal/rules"
)

type caso struct {
	ID    string `json:"id"`
	Valor string `json:"valor"`
}

type entrada struct {
	Casos []caso `json:"casos"`
}

type salida struct {
	Conteos    map[string]automata.Stats `json:"conteos"`
	Veredictos []bool                    `json:"veredictos"`
}

func main() {
	var in entrada
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail("entrada inválida: %v", err)
	}

	conteos := make(map[string]automata.Stats, 15)
	for _, r := range rules.All() {
		if r.TieneER {
			conteos[r.ID] = r.Counts()
		}
	}

	veredictos := make([]bool, len(in.Casos))
	for i, c := range in.Casos {
		r := rules.Get(c.ID)
		if r == nil || !r.TieneER {
			fail("regla desconocida: %q", c.ID)
		}
		veredictos[i] = r.Accepts(c.Valor)
	}

	if err := json.NewEncoder(os.Stdout).Encode(salida{Conteos: conteos, Veredictos: veredictos}); err != nil {
		fail("salida inválida: %v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
