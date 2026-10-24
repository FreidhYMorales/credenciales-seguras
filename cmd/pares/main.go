// Command pares regenerates testdata/pares.json, or prints the same 20 pairs as
// the Markdown tables of the final document.
//
// The 10 valid pairs come from the same generator the API uses, with recorded
// seeds, so the file can be rebuilt byte for byte. The 10 invalid pairs are
// authored by hand; each one has exactly one defective credential and names the
// rule it must break.
//
// The construction itself lives in the testdata package, so the regeneration
// test can call the same code instead of duplicating it.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/FreidhYMorales/credenciales-seguras/testdata"
)

func main() {
	md := flag.Bool("md", false, "imprime los pares como tablas Markdown en vez de escribir pares.json")
	flag.Parse()

	pairs, err := testdata.Build()
	if err != nil {
		log.Fatalf("construcción: %v", err)
	}
	if err := testdata.Verify(pairs); err != nil {
		log.Fatalf("verificación: %v", err)
	}

	if *md {
		// The document includes this fragment, so it goes to stdout and never
		// to the log, which would prefix every table with a timestamp.
		if _, err := os.Stdout.Write(testdata.Markdown(pairs)); err != nil {
			log.Fatalf("escritura: %v", err)
		}
		return
	}

	raw, err := testdata.Encode(pairs)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("testdata/pares.json", raw, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("testdata/pares.json: %d pares (%d válidos, %d inválidos)",
		len(pairs), testdata.ValidPairs, len(pairs)-testdata.ValidPairs)
}
