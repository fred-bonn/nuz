package web

import "embed"

// Templates holds the HTML page templates, one file per page.
//
//go:embed templates/*.html
var Templates embed.FS
