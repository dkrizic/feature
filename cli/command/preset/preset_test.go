package preset

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestPreSet_InvalidEndpoint(t *testing.T) {
	// Test that PreSet handles connection errors gracefully
	
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
	
	// The command expects "key" and "value" arguments
	// Since we're using an invalid endpoint, we expect a connection error
	err := PreSet(context.Background(), cmd)
	
	// We expect an error because the connection will fail
	assert.Error(t, err, "PreSet should return an error with invalid endpoint")
}
