package service

import (
	"context"
	"embed"
	"html/template"
	"log/slog"

	"go.opentelemetry.io/otel"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

// ParseTemplates parses all embedded templates and returns a *template.Template.
// Logs any parse errors using slog.
func ParseTemplates(ctx context.Context) *template.Template {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "ParseTemplates")
	defer span.End()

	tmpl, err := template.ParseFS(templateFS, "templates/*.gohtml")
	if err != nil {
		slog.Error("Failed to parse templates", "error", err)
		return nil
	}
	return tmpl
}
