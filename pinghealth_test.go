package main

import (
	"testing"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/assert"
)

// skipIfICMPUnavailable skips the test when the environment doesn't permit
// sending ICMP echo requests (e.g. CI runners without net.ipv4.ping_group_range
// configured for unprivileged ping), so failures here don't mask real bugs.
func skipIfICMPUnavailable(t *testing.T) {
	t.Helper()
	pinger, err := probing.NewPinger("localhost")
	if err != nil {
		t.Skipf("skipping: cannot create pinger: %s", err)
	}
	pinger.Count = 1
	pinger.Timeout = time.Second
	if err := pinger.Run(); err != nil {
		t.Skipf("skipping: ICMP ping not permitted in this environment: %s", err)
	}
}

func TestWorkingPingCheck(t *testing.T) {
	skipIfICMPUnavailable(t)
	h := NewPingCheck([]string{"localhost"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}

func TestFailLastPingCheck(t *testing.T) {
	skipIfICMPUnavailable(t)
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"localhost", "192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "")
}

func TestFailFirstPingCheck(t *testing.T) {
	skipIfICMPUnavailable(t)
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"192.0.2.0", "localhost"})
	result := h.Check()
	assert.Equal(t, result.healthy, true)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}

func TestFailingPingCheck(t *testing.T) {
	skipIfICMPUnavailable(t)
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"192.0.2.0"})
	result := h.Check()
	assert.Equal(t, result.healthy, false)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}
func TestPingChecksAreRepeteable(t *testing.T) {
	skipIfICMPUnavailable(t)
	// "192.0.2.0" is a test ip according to https://www.rfc-editor.org/rfc/rfc5737#section-3
	h := NewPingCheck([]string{"192.0.2.0"})
	result := h.Check()
	result = h.Check()
	assert.Equal(t, result.healthy, false)
	assert.Equal(t, result.err, "")
	assert.Equal(t, result.output, "192.0.2.0: 1 packets transmitted, 0 packets received, 100% packet loss, ")
}
