package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	advertised   = false // advertised holds a bool value to show whether the service ip is bgp advertised
	flagConfig   = flag.String("config", "/etc/bgp-lb/config.json", "Config file path")
	flagLogLevel = flag.String("log-level", "info", "Log level (debug|info|warning|error)")
)

func initLogger(logLevel string) {
	log.SetFormatter(&log.TextFormatter{})

	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.WithFields(log.Fields{
			"level": logLevel}).Fatal("Unsupported log level")
	}
}
func main() {
	flag.Parse()
	initLogger(*flagLogLevel)
	config, err := readConfig(*flagConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Start bgp server
	bgp, err := initBgpServer(
		config.Bgp.Local.RouterId,
		config.Bgp.Local.AS,
		config.Bgp.Local.ListenPort,
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, peer := range config.Bgp.Peers {
		if err := bgp.AddPeer(peer.Address, peer.AS); err != nil {
			log.Fatal(err)
		}
	}

	// TODO: Set up ipvs service here(?). Otherwise this needs to be set as
	// a systemd service ExecStartPre.
	// TODO: Add a dummy interface for the service ip (?). Otherwise this
	// needs to be set as a systemd service ExecStartPre.

	// Set up the healthcheck. TODO: Make an interface and allow different
	// kinds of healthchecks
	h := NewHttpCheck(
		config.Service.HttpHealthCheck.Name,
		fmt.Sprintf("%s http server check", config.Service.HttpHealthCheck.Name),
		config.Service.HttpHealthCheck.Url,
	)

	for t := time.Tick(time.Second * time.Duration(1)); ; <-t {
		status := h.Check()
		if status.err != "" {
			log.Warn(fmt.Sprintf("Healthcheck error: %s", status.err))
		}
		if status.healthy && !advertised {
			ServiceOn(bgp, config)
		}
		if !status.healthy && advertised {
			if status.output != "" {
				log.Warn(fmt.Sprintf("Healthcheck failed: %s", status.output))
			}
			ServiceOff(bgp, config)
		}
	}
}

func ServiceOn(bgp *BgpServer, config *config) {
	if err := bgp.AddV4Path(
		config.Service.IP,
		32,
		config.Bgp.Local.RouterId,
	); err != nil {
		log.Fatal(err)
	}
	bgp.ListV4Paths()
	advertised = true
}

func ServiceOff(bgp *BgpServer, config *config) {
	if err := bgp.DeleteV4Path(
		config.Service.IP,
		32,
		config.Bgp.Local.RouterId,
	); err != nil {
		log.Fatal(err)
	}
	advertised = false
	bgp.ListV4Paths()
}
