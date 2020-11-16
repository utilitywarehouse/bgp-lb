package main

import (
	"fmt"
	libipvs "github.com/moby/ipvs"
	"net"
	"strings"
	"syscall"
)

// addIPVSService adds a new ipvs service based on ip, protocol and listen port
func addIPVSService(ip, proto string, port uint16) (*libipvs.Service, error) {
	h, err := libipvs.New("")
	if err != nil {
		return nil, fmt.Errorf("IPVS interface can't be initialized: %v", err)
	}
	defer h.Close()
	svc := toIPVSService(ip, proto, port)
	return svc, h.NewService(svc)
}

// addIPVSDestination add a destination (ip and port) under a service
func addIPVSDestination(svc *libipvs.Service, ip string, port uint16) error {
	h, err := libipvs.New("")
	if err != nil {
		return fmt.Errorf("IPVS interface can't be initialized: %v", err)
	}
	defer h.Close()
	return h.NewDestination(svc, toIPVSDestination(ip, port))
}

// cleanIPVSServices deletes all ipvs services with the given ip
func cleanIPVSServices(ip string) error {
	h, err := libipvs.New("")
	if err != nil {
		return fmt.Errorf("IPVS interface can't be initialized: %v", err)
	}
	defer h.Close()
	svcs, err := h.GetServices()
	if err != nil {
		return fmt.Errorf("Cannot retrieve ipvs services: %v", err)
	}
	cleanIP := net.ParseIP(ip)
	for _, svc := range svcs {
		if svc.Address.Equal(cleanIP) {
			if err := h.DelService(svc); err != nil {
				return fmt.Errorf("Cannot delete ipvs svc: %v", err)
			}
		}
	}
	return nil
}

// toIPVSService converts ip, protocol and port to the equivalent IPVS Service
// structure.
func toIPVSService(ip, proto string, port uint16) *libipvs.Service {
	return &libipvs.Service{
		Address:       net.ParseIP(ip),
		Protocol:      stringToProtocol(proto),
		Port:          port,
		SchedName:     "rr",
		AddressFamily: syscall.AF_INET,
		Netmask:       0xffffffff,
	}
}

// toIPVSDestination converts a RealServer to the equivalent IPVS Destination structure.
func toIPVSDestination(ip string, port uint16) *libipvs.Destination {
	return &libipvs.Destination{
		Address: net.ParseIP(ip),
		Port:    port,
		Weight:  1,
	}
}

// stringToProtocolType returns the protocol type for the given name
func stringToProtocol(protocol string) uint16 {
	switch strings.ToLower(protocol) {
	case "tcp":
		return uint16(syscall.IPPROTO_TCP)
	case "udp":
		return uint16(syscall.IPPROTO_UDP)
	case "sctp":
		return uint16(syscall.IPPROTO_SCTP)
	}
	return uint16(0)
}
