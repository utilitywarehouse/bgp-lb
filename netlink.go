package main

import (
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

// ensureServiceDevice looks for a device of the given name and creates a dummy
// one if not found
func ensureServiceDevice(name string) error {
	h := netlink.Handle{}
	defer h.Delete()
	_, err := h.LinkByName(name)
	if err != nil {
		_, notFound := err.(netlink.LinkNotFoundError)
		if notFound {
			d := &netlink.Dummy{netlink.LinkAttrs{
				Name: name,
			}}
			return netlink.LinkAdd(d)
		}
		return err
	}
	return nil
}

// ensureDeviceUp bring up a netlink device
func ensureDeviceUp(name string) error {
	h := netlink.Handle{}
	defer h.Delete()
	link, err := h.LinkByName(name)
	if err != nil {
		return err
	}
	return h.LinkSetUp(link)
}

// flushIPv4Addresses deletes all the ipv4 addresses from a device
func flushIPv4Addresses(device string) error {
	h := netlink.Handle{}
	defer h.Delete()
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
func addAddressToDevice(ip, device string) error {
	h := netlink.Handle{}
	defer h.Delete()
	link, err := h.LinkByName(device)
	if err != nil {
		return err
	}
	ipv4Addr := net.ParseIP(ip)
	ipv4Mask := net.CIDRMask(32, 32) // default to /32 service addresses
	return h.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipv4Addr,
			Mask: ipv4Mask,
		},
	})
}
