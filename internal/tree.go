package internal

import (
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

	// Sort roots and children for consistent output
	sortInterfaces(roots, opts)
	for _, iface := range interfaces {
		if len(iface.Children) > 0 {
			sortInterfaces(iface.Children, opts)
		}
	}

	return roots
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
