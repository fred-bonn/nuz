package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/fred-bonn/nuz/web"
)

var pages = template.Must(template.ParseFS(web.Templates, "templates/*.html"))

func renderPage(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := pages.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
