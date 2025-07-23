package main

import (
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type BaselineCheck struct {
	pingers []*probing.Pinger
}

func NewBaselineCheck() BaselineCheck {
	// By default, check baseline connectivity with three well-known dns providers
	addresses := []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}
	return NewBaselineCheckWithAddresses(addresses)
}

func NewBaselineCheckWithAddresses(addresses []string) BaselineCheck {
	check := BaselineCheck{}
	for _, address := range addresses {
		pinger, err := probing.NewPinger(address)
		if err != nil {
			panic(err)
		}
		pinger.Count = 1
		pinger.Timeout = 5 * time.Second
		check.pingers = append(check.pingers, pinger)
	}
	return check
}

func (pc BaselineCheck) Check() Result {
	healthy := false
	errMsg := ""
	out := ""

	// If any pinger succeeds, consider that connectivity is healthy
	for _, pinger := range pc.pingers {
		err := pinger.Run() // Blocks until finished.
		if err != nil {
			errMsg += fmt.Sprintf("ping probe for %v failed to run with error: %s, ", pinger.Addr(), err)
			continue
		}
		stats := pinger.Statistics()
		if stats.PacketLoss == 0 {
			healthy = true
		} else {
			out += fmt.Sprintf("%v: %d packets transmitted, %d packets received, %v%% packet loss, ",
				pinger.Addr(), stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		}
	}

	return Result{
		healthy: healthy,
		err:     errMsg,
		output:  out,
	}
}
