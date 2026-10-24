//go:build gui && (linux || freebsd || openbsd || netbsd)

package main

import (
	"os"
	"os/exec"
)

// windowAvailable reports whether there is a display to open the GTK window
// on: without one webview_create returns NULL and webview_go would dereference
// it.
func windowAvailable() bool {
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

// openInBrowser hands the URL to the desktop's default browser.
func openInBrowser(url string) error {
	return exec.Command("xdg-open", url).Start()
}
