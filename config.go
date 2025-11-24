package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// config includes all the config
type config struct {
	Bgp     bgpConfig     `json:"bgp"`
	Service serviceConfig `json:"service"`
}

// bgpConfig includes config for bgp peers and the local bgp server
type bgpConfig struct {
	Peers []peerConfig `json:"peers"`
	Local localConfig  `json:"local"`
}

// peerConfig contains the config for a bgp peer
type peerConfig struct {
	Address string `json:"address"`
	AS      uint32 `json:"as"`
}

// localConfig contains the bgp config for the local server
type localConfig struct {
	RouterId   string `json:"routerID"`
	AS         uint32 `json:"as"`
	ListenPort int32  `json:"listenPort"`
}

// serviceConfig contains the advertised service ip and the healthcheck
type serviceConfig struct {
	Name            string                 `json:"name"`
	IP              string                 `json:"ip"`
	PrefixLength    int                    `json:"prefixLength"`
	Ports           []servicePortConfig    `json:"ports"`
	Protocol        string                 `json:"protocol"`
	HttpHealthCheck *httpHealthCheckConfig `json:"httphealthcheck"`
	PingHealthCheck *pingHealthCheckConfig `json:"pinghealthcheck"`
}

// servicePortsConfig contains the mapping between a service and a local port
type servicePortConfig struct {
	ServicePort uint16 `json:"servicePort"`
	TargetPort  uint16 `json:"targetLocalPort"`
}

// httpHealthCheckConfig contains the local port the http health endpoint listens to
type httpHealthCheckConfig struct {
	Path string `json:"path"`
	Port int    `json:"port"`
}

// pingHealthCheckConfig contains the address for the pinger to check
type pingHealthCheckConfig struct {
	Addresses []string `json:"addresses"`
}

func readConfig(path string) (*config, error) {
	// Default service prefix to /32 to avoid using /0 if omitted from the
	// config file
	conf := &config{
		Service: serviceConfig{
			PrefixLength: 32,
		},
	}
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return conf, fmt.Errorf("error reading config file: %v", err)
	}
	if err = json.Unmarshal(fileContent, conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}
	return conf, nil
}
