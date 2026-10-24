//go:build gui && windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// windowAvailable reports whether the WebView2 runtime is installed. Without
// it webview cannot create a window at all, and the failure would be silent
// because the binary is linked as a GUI application.
func windowAvailable() bool {
	bases := []string{
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("ProgramFiles"),
		os.Getenv("LOCALAPPDATA"),
	}
	for _, base := range bases {
		if base == "" {
			continue
		}
		pattern := filepath.Join(base, "Microsoft", "EdgeWebView", "Application", "*", "msedgewebview2.exe")
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// openInBrowser hands the URL to the default browser.
func openInBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
