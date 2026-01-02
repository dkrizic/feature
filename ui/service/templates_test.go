package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTemplates(t *testing.T) {
	ctx := context.Background()
	tmpl := ParseTemplates(ctx)

	assert.NotNil(t, tmpl, "ParseTemplates should return a non-nil template")

	// Check if the templates have been parsed
	// The actual templates should be embedded in the binary
	if tmpl != nil {
		// Templates should have been loaded from the embedded FS
		assert.NotNil(t, tmpl.Lookup("index.gohtml"), "index.gohtml template should exist")
		assert.NotNil(t, tmpl.Lookup("features_list.gohtml"), "features_list.gohtml template should exist")
	}
}
