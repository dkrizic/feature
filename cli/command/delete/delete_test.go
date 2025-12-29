package delete

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestDelete_InvalidEndpoint(t *testing.T) {
	// Test that Delete handles connection errors gracefully
	
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
	
	// The command expects a "key" argument
	// Since we're using an invalid endpoint, we expect a connection error
	err := Delete(context.Background(), cmd)
	
	// We expect an error because the connection will fail
	assert.Error(t, err, "Delete should return an error with invalid endpoint")
}
