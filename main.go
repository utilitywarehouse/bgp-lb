package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	advertised       = false // advertised holds a bool value to show whether the service ip is bgp advertised
	flagConfig       = flag.String("config", "/etc/bgp-lb/config.json", "Config file path")
	flagLogLevel     = flag.String("log-level", "info", "Log level (debug|info|warning|error)")
	flagNetworkSetup = flag.Bool("network-setup", true, "Whether to set up a net interface for the service address on the host")
	flagIPVSSetup    = flag.Bool("ipvs-setup", false, "Will flush IPVS table and add a route from the service address to the target host port. Effective only when combined with -network-setup")
	flagMetricsAddr  = flag.String("metrics-address", ":8081", "Metrics server address")
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
	if *flagNetworkSetup {
		netlinkSetup(config.Service, config.Bgp.Local.RouterId, *flagIPVSSetup)
	}
	go startMetricsServer(*flagMetricsAddr)
	// init metric with 0 value, in case healthcheck fails
	unsetBGPPathAdvertisementMetric(config.Service.IP, fmt.Sprint(config.Service.PrefixLength), config.Bgp.Local.RouterId)

	h := healthCheckSetup(config.Service)
	for t := time.Tick(time.Second * time.Duration(1)); ; <-t {
		log.Debug("Running a new healthcheck")
		res := h.Check()
		if res.err != "" {
			log.Warn(fmt.Sprintf("Healthcheck error: %s\n", res.err))
		}
		if res.healthy {
			log.Debug("Healthcheck succeeded")
		} else {
			if res.output != "" {
				log.Warn(fmt.Sprintf("Healthcheck failed: %s\n", res.output))
			} else {
				log.Warn("Healthcheck failed")
			}
		}
		if res.healthy && !advertised {
			ServiceOn(bgp, config)
		}
		if !res.healthy && advertised {
			ServiceOff(bgp, config)
		}
	}
}

func ServiceOn(bgp *BgpServer, config *config) {
	if err := bgp.AddV4Path(
		config.Service.IP,
		uint32(config.Service.PrefixLength),
		config.Bgp.Local.RouterId,
	); err != nil {
		log.Fatal(err)
	}
	bgp.ListV4Paths()
	advertised = true
	log.Info("Service on")
}

func ServiceOff(bgp *BgpServer, config *config) {
	if err := bgp.DeleteV4Path(
		config.Service.IP,
		uint32(config.Service.PrefixLength),
		config.Bgp.Local.RouterId,
	); err != nil {
		log.Fatal(err)
	}
	bgp.ListV4Paths()
	advertised = false
	log.Info("Service off")
}
