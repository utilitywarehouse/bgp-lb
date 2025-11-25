package main

import (
	"net"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// ensureServiceDevice looks for a device of the given name and creates a dummy
// one if not found
func ensureServiceDevice(name string) error {
	h := netlink.Handle{}
	defer h.Close()
	_, err := h.LinkByName(name)
	if err != nil {
		_, notFound := err.(netlink.LinkNotFoundError)
		if notFound {
			d := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{
				Name: name,
			}}
			return netlink.LinkAdd(d)
		}
		return err
	}
	return nil
}

// flushIPv4Addresses deletes all the ipv4 addresses from a device
func flushIPv4Addresses(device string) error {
	h := netlink.Handle{}
	defer h.Close()
	link, err := h.LinkByName(device)
	if err != nil {
		return err
	}
	addrs, err := h.AddrList(link, syscall.AF_INET)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		if err := h.AddrDel(link, &addr); err != nil {
			return err
		}
	}
	return nil
}

// addAddressToDevice adds an ip address to a device
func addAddressToDevice(ip, device string, prefixLength int) error {
	h := netlink.Handle{}
	defer h.Close()
	link, err := h.LinkByName(device)
	if err != nil {
		return err
	}
	ipv4Addr := net.ParseIP(ip)
	ipv4Mask := net.CIDRMask(prefixLength, 32)
	return h.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipv4Addr,
			Mask: ipv4Mask,
		},
	})
}

// netlinkSetup applies the needed host network configuration based on the
// service config
func netlinkSetup(serviceConfig serviceConfig, localIP string, setupIPVS bool) {
	// Ensure the dummy device exists
	if err := ensureServiceDevice(serviceConfig.Name); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot ensure service link device")
	}
	// Add the service ip after cleaning all pre-existing ipv4 addresses
	if err := flushIPv4Addresses(serviceConfig.Name); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"device": serviceConfig.Name,
		}).Fatal("Failed to clean ipv4 addresses from device")
	}
	if err := addAddressToDevice(serviceConfig.IP, serviceConfig.Name, serviceConfig.PrefixLength); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot add address to service link device")
	}
	// If setting IPVS is not required, we are done here
	if !setupIPVS {
		return
	}
	// Ensure ipvs service and add the local router as destination
	if err := cleanIPVSServices(serviceConfig.IP); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot clean existing ipvs services")
	}
	for _, port := range serviceConfig.Ports {
		svc, err := addIPVSService(serviceConfig.IP, serviceConfig.Protocol, port.ServicePort)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Cannot add ipvs service")
		}
		if err := addIPVSDestination(svc, localIP, port.TargetPort); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Cannot add ipvs service destination")
		}
	}
}
