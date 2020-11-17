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
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to read config file")
	}

	bgp := bgpSetup(config.Bgp)
	netlinkSetup(config.Service, config.Bgp.Local.RouterId)
	h := healthCheckSetup(config.Service)

	for t := time.Tick(time.Second * time.Duration(1)); ; <-t {
		res := h.Check()
		if res.err != "" {
			log.Warn(fmt.Sprintf("Healthcheck error: %s", res.err))
		}
		if res.healthy && !advertised {
			ServiceOn(bgp, config)
		}
		if !res.healthy && advertised {
			if res.output != "" {
				log.Warn(fmt.Sprintf("Healthcheck failed: %s", res.output))
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
