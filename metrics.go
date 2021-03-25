package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	bgpPathAdvertisement = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bgp_lb_path_advertisement",
		Help: "Info about whether a path is advertised via the bgp daemon. It can be 0 or 1.",
	},
		[]string{
			"prefix",
			"prefix_length",
			"next_hop",
		},
	)
)

func init() {
	prometheus.MustRegister(bgpPathAdvertisement)
}

func setBGPPathAdvertisementMetric(prefix, prefixLen, nexthop string) {
	bgpPathAdvertisement.With(prometheus.Labels{
		"prefix":        prefix,
		"prefix_length": prefixLen,
		"next_hop":      nexthop,
	}).Set(1)
}

func unsetBGPPathAdvertisementMetric(prefix, prefixLen, nexthop string) {
	bgpPathAdvertisement.With(prometheus.Labels{
		"prefix":        prefix,
		"prefix_length": prefixLen,
		"next_hop":      nexthop,
	}).Set(0)
}

func startMetricsServer(listenAddress string) {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
