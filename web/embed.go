// Package web embeds the frontend: no build step, served by the same binary.
package web

import "embed"

// FS holds index.html, app.js and styles.css.
//
//go:embed index.html app.js styles.css
var FS embed.FS
