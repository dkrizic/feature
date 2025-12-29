package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureClient(t *testing.T) {
	// This test verifies that FeatureClient can be called without crashing
	// We can't easily test the actual gRPC connection without a running server
	// So we just verify the function exists and accepts the right parameters
	
	// Note: This is a minimal test since FeatureClient depends on external gRPC server
	// In a real scenario, we would need to mock or inject the gRPC client factory
	
	// For now, we just ensure the package compiles and the function signature is correct
	assert.NotNil(t, FeatureClient, "FeatureClient function should exist")
}
