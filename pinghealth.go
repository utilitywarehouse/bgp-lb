package main

import (
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type PingCheck struct {
	addresses []string
}

func NewPingCheck(addresses []string) PingCheck {
	return PingCheck{addresses: addresses}
}

func (pc PingCheck) Check() Result {
	healthy := false
	errMsg := ""
	out := ""

	// If any pinger succeeds, consider that connectivity is healthy
	for _, address := range pc.addresses {
		pinger, err := probing.NewPinger(address)
		if err != nil {
			errMsg += fmt.Sprintf("%v: failed to create probe with error: %s, ", address, err)
			continue
		}
		pinger.Count = 1
		pinger.Timeout = 5 * time.Second

		err = pinger.Run() // Blocks until finished.
		if err != nil {
			errMsg += fmt.Sprintf("%v: failed to run probe with error: %s, ", address, err)
			continue
		}

		stats := pinger.Statistics()
		if stats.PacketLoss == 0 {
			healthy = true
			break
		} else {
			out += fmt.Sprintf("%v: %d packets transmitted, %d packets received, %v%% packet loss, ",
				address, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		}
	}

	return Result{
		healthy: healthy,
		err:     errMsg,
		output:  out,
	}
}
