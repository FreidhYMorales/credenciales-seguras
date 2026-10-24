//go:build !gui

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// defaultAddr keeps the documented port when there is no window to open.
func defaultAddr() string { return "127.0.0.1:8080" }

// openWindow has no window to open without the gui tag: it prints the address
// and waits for Ctrl+C.
func openWindow(url string) error {
	log.Printf("abrí %s en el navegador (Ctrl+C para salir)", url)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	return nil
}
