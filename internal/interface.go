package internal

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/vishvananda/netlink"
)

// Interface represents a network interface with all its properties
type Interface struct {
	Name       string
	Type       string
	State      string
	IPs        []net.IP
	IPv4Addrs  []net.IP
	IPv6Addrs  []net.IP
	IPv4Nets   []*net.IPNet // IPv4 addresses with CIDR
	IPv6Nets   []*net.IPNet // IPv6 addresses with CIDR
	MAC        net.HardwareAddr
	MTU        int
	Master     string
	MasterIdx  int
	Children   []*Interface
	Namespace  string
	Peer       string
	PeerNS     string
	Driver     string
	Model      string
	Speed      int64
	RxBytes    uint64
	RxPackets  uint64
	TxBytes    uint64
	TxPackets  uint64
	Index      int
	LinkIndex  int
	IsPhysical bool
}

// Options holds the runtime configuration
type Options struct {
	ShowAll       bool
	AllNamespaces bool
	DirectionDown bool
	Columns       []string
	ListFormat    bool
	JSONOutput    bool
	NoHeadings    bool
	ExcludeTypes  []string
	FilterTypes   []string
}

// DiscoverInterfaces discovers all network interfaces in the current namespace
func DiscoverInterfaces(opts *Options) ([]*Interface, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}

	interfaces := make([]*Interface, 0, len(links))

	for _, link := range links {
		attrs := link.Attrs()

		// Skip if interface is down and we're not showing all
		if !opts.ShowAll && attrs.Flags&net.FlagUp == 0 {
			continue
		}

		iface := &Interface{
			Name:      attrs.Name,
			Type:      getLinkType(link),
			State:     getState(attrs.Flags),
			MAC:       attrs.HardwareAddr,
			MTU:       attrs.MTU,
			MasterIdx: attrs.MasterIndex,
			Index:     attrs.Index,
			Namespace: "(root)", // TODO: get actual namespace name
		}

		// Get master device name if exists
		if attrs.MasterIndex > 0 {
			if master, err := netlink.LinkByIndex(attrs.MasterIndex); err == nil {
				iface.Master = master.Attrs().Name
			}
		}

		// Get link index for VLANs and veth
		if attrs.ParentIndex > 0 {
			iface.LinkIndex = attrs.ParentIndex
		}

		// Get veth peer if it's a veth interface
		if iface.Type == "veth" {
			if peer, peerNS := getVethPeer(link); peer != "" {
				iface.Peer = peer
				iface.PeerNS = peerNS
			}
		}

		// Get IP addresses
		if addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL); err == nil {
			for _, addr := range addrs {
				ip := addr.IPNet.IP
				iface.IPs = append(iface.IPs, ip)
				if ip.To4() != nil {
					iface.IPv4Addrs = append(iface.IPv4Addrs, ip)
					iface.IPv4Nets = append(iface.IPv4Nets, addr.IPNet)
				} else {
					iface.IPv6Addrs = append(iface.IPv6Addrs, ip)
					iface.IPv6Nets = append(iface.IPv6Nets, addr.IPNet)
				}
			}
		}

		// Get statistics
		if stats := attrs.Statistics; stats != nil {
			iface.RxBytes = stats.RxBytes
			iface.RxPackets = stats.RxPackets
			iface.TxBytes = stats.TxBytes
			iface.TxPackets = stats.TxPackets
		}

		// Get driver and model from sysfs
		iface.Driver = getDriver(attrs.Name)
		iface.Speed = getSpeed(attrs.Name)

		// Determine if physical
		iface.IsPhysical = isPhysicalInterface(iface)

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// getLinkType returns the interface type
func getLinkType(link netlink.Link) string {
	switch link.(type) {
	case *netlink.Device:
		return "ether"
	case *netlink.Bridge:
		return "bridge"
	case *netlink.Vlan:
		return "vlan"
	case *netlink.Veth:
		return "veth"
	case *netlink.Bond:
		return "bond"
	case *netlink.Dummy:
		return "dummy"
	case *netlink.Tuntap:
		tun := link.(*netlink.Tuntap)
		if tun.Mode == netlink.TUNTAP_MODE_TUN {
			return "tun"
		}
		return "tap"
	case *netlink.Wireguard:
		return "wireguard"
	default:
		// Try to get type from attrs
		attrs := link.Attrs()
		if attrs.EncapType == "ether" {
			return "ether"
		}
		return "unknown"
	}
}

// getState returns the interface state
func getState(flags net.Flags) string {
	if flags&net.FlagUp != 0 {
		return "UP"
	}
	return "DOWN"
}

// getVethPeer attempts to get the veth peer interface name
func getVethPeer(link netlink.Link) (string, string) {
	veth, ok := link.(*netlink.Veth)
	if !ok {
		return "", ""
	}

	attrs := veth.Attrs()
	if attrs.ParentIndex > 0 {
		// Try to get peer by index
		if peer, err := netlink.LinkByIndex(attrs.ParentIndex); err == nil {
			return peer.Attrs().Name, "" // TODO: detect namespace
		}
	}

	return "", ""
}

// getDriver reads the driver name from sysfs
func getDriver(ifname string) string {
	driverPath := filepath.Join("/sys/class/net", ifname, "device", "driver")
	if target, err := os.Readlink(driverPath); err == nil {
		return filepath.Base(target)
	}
	return ""
}

// getSpeed reads the link speed from sysfs
func getSpeed(ifname string) int64 {
	speedPath := filepath.Join("/sys/class/net", ifname, "speed")
	data, err := os.ReadFile(speedPath)
	if err != nil {
		return 0
	}

	var speed int64
	fmt.Sscanf(string(data), "%d", &speed)
	return speed
}

// isPhysicalInterface determines if an interface is physical
func isPhysicalInterface(iface *Interface) bool {
	// Physical interfaces typically have:
	// - A driver
	// - Are of type ether or wlan
	// - No master device
	// - Not virtual types

	virtualTypes := map[string]bool{
		"bridge": true,
		"vlan":   true,
		"veth":   true,
		"bond":   true,
		"dummy":  true,
		"tun":    true,
		"tap":    true,
	}

	if virtualTypes[iface.Type] {
		return false
	}

	// Has driver suggests physical
	if iface.Driver != "" {
		return true
	}

	// Loopback is not physical
	if iface.Name == "lo" {
		return false
	}

	return iface.Type == "ether"
}

// FilterInterfaces applies type filtering
func FilterInterfaces(interfaces []*Interface, opts *Options) []*Interface {
	if len(opts.FilterTypes) == 0 && len(opts.ExcludeTypes) == 0 {
		return interfaces
	}

	filtered := make([]*Interface, 0, len(interfaces))

	for _, iface := range interfaces {
		// Check if type should be filtered
		if len(opts.FilterTypes) > 0 {
			include := false
			for _, t := range opts.FilterTypes {
				if strings.EqualFold(iface.Type, t) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		// Check if type should be excluded
		if len(opts.ExcludeTypes) > 0 {
			exclude := false
			for _, t := range opts.ExcludeTypes {
				if strings.EqualFold(iface.Type, t) {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		filtered = append(filtered, iface)
	}

	return filtered
}
