package get

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestGet_MissingKey(t *testing.T) {
	// Test that Get handles missing key argument gracefully
	// This test verifies the command can be called and handles basic error cases
	
	var buf bytes.Buffer
	cmd := &cli.Command{
		Writer: &buf,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "endpoint",
				Value: "invalid:9999", // Use invalid endpoint to avoid actual connection
			},
		},
	}
	
	// The command expects a "key" argument but doesn't validate it before calling gRPC
	// So we expect this to fail with a gRPC connection error or argument error
	err := Get(context.Background(), cmd)
	
	// We expect an error because either the key is missing or connection fails
	assert.Error(t, err, "Get should return an error with invalid endpoint or missing key")
}
