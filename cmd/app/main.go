// Command app serves the credential validator and generator.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/FreidhYMorales/credenciales-seguras/internal/api"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("credenciales seguras en http://%s", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
