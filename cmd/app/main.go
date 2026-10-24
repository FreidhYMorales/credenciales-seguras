// Command app serves the credential validator and generator: as a native
// desktop window when built with `-tags gui`, otherwise in the browser.
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/FreidhYMorales/credenciales-seguras/internal/api"
)

func main() {
	addr := flag.String("addr", defaultAddr(), `listen address; "127.0.0.1:0" picks a free port`)
	flag.Parse()

	// Listen first so the window can be pointed at the real port, even when
	// the kernel picked it.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("no se pudo abrir %s: %v", *addr, err)
	}
	url := "http://" + ln.Addr().String() + "/"

	srv := &http.Server{
		Handler:           api.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("servidor detenido: %v", err)
		}
	}()

	// openWindow blocks until the user closes the window (or interrupts the
	// process when there is no window).
	if err := openWindow(url); err != nil {
		log.Printf("ventana: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("cierre: %v", err)
	}
}
