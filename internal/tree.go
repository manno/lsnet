package internal

import (
	"fmt"
	"os"
	"sort"
)

// BuildTree builds a hierarchical tree of interfaces
func BuildTree(interfaces []*Interface, opts *Options) []*Interface {
	// Create a map for quick lookup
	ifaceMap := make(map[string]*Interface)
	indexMap := make(map[int]*Interface)

	for _, iface := range interfaces {
		ifaceMap[iface.Name] = iface
		indexMap[iface.Index] = iface
	}

	// Build parent-child relationships
	roots := make([]*Interface, 0)

	for _, iface := range interfaces {
		hasParent := false

		// Check for master relationship (bridge/bond members)
		if iface.MasterIdx > 0 {
			if master, ok := indexMap[iface.MasterIdx]; ok {
				master.Children = append(master.Children, iface)
				hasParent = true
			}
		}

		// Check for link relationship (VLANs)
		if !hasParent && iface.LinkIndex > 0 && iface.Type == "vlan" {
			if parent, ok := indexMap[iface.LinkIndex]; ok {
				parent.Children = append(parent.Children, iface)
				hasParent = true
			}
		}

		// If no parent, it's a root
		if !hasParent {
			roots = append(roots, iface)
		}
	}

	// Discover listening ports
	listeningPorts, err := DiscoverListeningPorts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not discover listening ports: %v\n", err)
		listeningPorts = []ListeningPort{}
	}

	// Pre-index ports by IP for O(1) lookups when building IP nodes.
	// Wildcard-bound ports (0.0.0.0 / ::) are omitted from the index since
	// they cannot be attributed to a specific IP address.
	portsByIP := make(map[string][]ListeningPort)
	for _, p := range listeningPorts {
		if !isWildcardAddr(p.Address) {
			key := p.Address.String()
			portsByIP[key] = append(portsByIP[key], p)
		}
	}

	// Add IP addresses as child nodes for all interfaces
	for _, iface := range interfaces {
		addIPNodes(iface, portsByIP)
	}

	// Sort roots and children for consistent output
	sortInterfaces(roots, opts)
	for _, iface := range interfaces {
		if len(iface.Children) > 0 {
			sortInterfaces(iface.Children, opts)
		}
	}

	return roots
}

// addIPNodes adds IP addresses as child nodes of an interface
func addIPNodes(iface *Interface, portsByIP map[string][]ListeningPort) {
	// Add IPv4 addresses first, then IPv6
	for _, ipnet := range iface.IPv4Nets {
		ports := portsByIP[ipnet.IP.String()]
		ipNode := &Interface{
			Name:     ipnet.String(), // This includes CIDR notation
			Type:     "inet",
			State:    "",
			IsIPNode: true,
			Ports:    ports,
		}

		// Add port nodes as children of the IP node
		for _, port := range ports {
			portNode := &Interface{
				Name:  formatPort(port),
				Type:  port.Protocol,
				State: "",
			}
			ipNode.Children = append(ipNode.Children, portNode)
		}

		iface.Children = append(iface.Children, ipNode)
	}

	for _, ipnet := range iface.IPv6Nets {
		ports := portsByIP[ipnet.IP.String()]
		ipNode := &Interface{
			Name:     ipnet.String(), // This includes CIDR notation
			Type:     "inet6",
			State:    "",
			IsIPNode: true,
			Ports:    ports,
		}

		// Add port nodes as children of the IP node
		for _, port := range ports {
			portNode := &Interface{
				Name:  formatPort(port),
				Type:  port.Protocol,
				State: "",
			}
			ipNode.Children = append(ipNode.Children, portNode)
		}

		iface.Children = append(iface.Children, ipNode)
	}
}

// formatPort formats a listening port for display
func formatPort(port ListeningPort) string {
	return fmt.Sprintf("%d", port.Port)
}

// sortInterfaces sorts interfaces by physical-first, then by name
func sortInterfaces(interfaces []*Interface, opts *Options) {
	sort.Slice(interfaces, func(i, j int) bool {
		// Physical devices first (unless direction is down)
		if !opts.DirectionDown {
			if interfaces[i].IsPhysical != interfaces[j].IsPhysical {
				return interfaces[i].IsPhysical
			}
		}

		// Then sort by type priority
		iPrio := getTypePriority(interfaces[i].Type)
		jPrio := getTypePriority(interfaces[j].Type)
		if iPrio != jPrio {
			return iPrio < jPrio
		}

		// Finally by name
		return interfaces[i].Name < interfaces[j].Name
	})
}

// getTypePriority returns sort priority for interface types
func getTypePriority(iftype string) int {
	priorities := map[string]int{
		"ether":     1,
		"wlan":      2,
		"bridge":    3,
		"bond":      4,
		"vlan":      5,
		"veth":      6,
		"tun":       7,
		"tap":       8,
		"dummy":     9,
		"wireguard": 10,
		"inet":      100, // IP addresses come after interfaces
		"inet6":     101,
		"tcp":       200, // Ports come after IP addresses
		"udp":       201,
		"unknown":   99,
	}

	if prio, ok := priorities[iftype]; ok {
		return prio
	}
	return 50
}

// FlattenTree converts tree to flat list (for list mode)
func FlattenTree(roots []*Interface) []*Interface {
	result := make([]*Interface, 0)

	var walk func(*Interface)
	walk = func(iface *Interface) {
		result = append(result, iface)
		for _, child := range iface.Children {
			walk(child)
		}
	}

	for _, root := range roots {
		walk(root)
	}

	return result
}
