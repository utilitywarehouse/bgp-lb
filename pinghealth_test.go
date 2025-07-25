package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkingPingCheck(t *testing.T) {
	h := NewPingCheck([]string{"localhost"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}

func TestFailLastPingCheck(t *testing.T) {
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"localhost", "192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}

func TestFailFirstPingCheck(t *testing.T) {
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"192.0.2.0", "localhost"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}

func TestFailingPingCheck(t *testing.T) {
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, false)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}
