package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHealthCheck(t *testing.T) {
	h := healthCheckSetup(serviceConfig{})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}
