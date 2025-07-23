package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkingBaselineHealthCheck(t *testing.T) {
	h := NewBaselineCheckWithAddresses([]string{"localhost"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}

func TestPartiallyFailingBaselineHealthCheck(t *testing.T) {
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewBaselineCheckWithAddresses([]string{"localhost", "192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}

func TestFailingBaselineHealthCheck(t *testing.T) {
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewBaselineCheckWithAddresses([]string{"192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, false)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}
