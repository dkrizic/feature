package service

import (
	"embed"
	"html/template"
	"log/slog"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

// ParseTemplates parses all embedded templates and returns a *template.Template.
// Logs any parse errors using slog.
func ParseTemplates() *template.Template {
	tmpl, err := template.ParseFS(templateFS, "templates/*.gohtml")
	if err != nil {
		slog.Error("Failed to parse templates", "error", err)
		return nil
	}
	return tmpl
}
