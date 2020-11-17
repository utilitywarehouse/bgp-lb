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

func bgpSetup(config *config) *BgpServer {
	// Start bgp server
	bgp, err := initBgpServer(
		config.Bgp.Local.RouterId,
		config.Bgp.Local.AS,
		config.Bgp.Local.ListenPort,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot start bgp server")
	}
	// Add Peers
	for _, peer := range config.Bgp.Peers {
		if err := bgp.AddPeer(peer.Address, peer.AS); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Cannot add bgpp peer")
		}
	}
	return bgp
}

func netlinkSetup(config *config) {
	// Ensure ipvs service and add the local router as destination
	if err := cleanIPVSServices(config.Service.IP); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot clean existing ipvs services")
	}
	svc, err := addIPVSService(config.Service.IP, config.Service.Protocol, config.Service.ServicePort)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot add ipvs service")
	}
	if err := addIPVSDestination(svc, config.Bgp.Local.RouterId, config.Service.TargetPort); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot add ipvs service destination")
	}
	// Ensure the dummy device exists
	if err := ensureServiceDevice(config.Service.Name); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot ensure service link device")
	}
	// Add the service ip after cleaning all pre-existing ipv4 addresses
	if err := flushIPv4Addresses(config.Service.Name); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"device": config.Service.Name,
		}).Fatal("Failed to clean ipv4 addresses from device")
	}
	if err := addAddressToDevice(config.Service.IP, config.Service.Name); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot add address to service link device")
	}
}

func healthCheckSetup(config *config) Checker {
	return NewHttpCheck(config.Service.HttpHealthCheck.Port)
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

	bgp := bgpSetup(config)
	netlinkSetup(config)
	h := healthCheckSetup(config)

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
