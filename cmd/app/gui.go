//go:build gui

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	webview "github.com/webview/webview_go"
)

// defaultAddr asks the kernel for a free port: the window is the only client,
// so a fixed port would only clash with a second copy of the app.
func defaultAddr() string { return "127.0.0.1:0" }

// openWindow shows the application in a native window and returns when the
// user closes it. Without a web engine (or without a display) it falls back to
// the browser instead of failing silently.
func openWindow(url string) error {
	if !windowAvailable() {
		log.Printf("sin motor de ventana; abro %s en el navegador", url)
		if err := openInBrowser(url); err != nil {
			log.Printf("navegador: %v", err)
		}
		return waitInterrupt()
	}

	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Credenciales seguras")
	w.SetSize(1280, 860, webview.HintNone)
	// The UI is a local server, so the window is just a viewer: the pages
	// never navigate away from 127.0.0.1.
	w.Navigate(url)
	log.Printf("ventana abierta en %s", url)
	w.Run()
	return nil
}

// waitInterrupt blocks until Ctrl+C or SIGTERM.
func waitInterrupt() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	return nil
}
