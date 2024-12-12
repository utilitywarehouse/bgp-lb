package main

import (
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type PingCheck struct {
	pinger *probing.Pinger
}

func NewPingCheck(address string) PingCheck {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		panic(err)
	}
	pinger.Count = 1
	pinger.Timeout = 5 * time.Second
	return PingCheck{
		pinger: pinger,
	}
}

func (pc PingCheck) Check() Result {
	err := pc.pinger.Run() // Blocks until finished.
	healthy := err == nil
	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf("Pinging failed: %v", err)
	}
	stats := pc.pinger.Statistics()
	out := fmt.Sprintf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)

	return Result{
		healthy: healthy,
		err:     errMsg,
		output:  out,
	}
}
