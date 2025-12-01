package internal

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

// OutputTree prints the interface tree
func OutputTree(roots []*Interface, opts *Options) error {
	if opts.JSONOutput {
		return outputJSON(roots)
	}

	cols := NormalizeColumns(opts.Columns)

	// Print headers
	if !opts.NoHeadings {
		printHeaders(cols)
	}

	// Print tree
	for _, root := range roots {
		printInterface(root, "", true, cols, opts)
	}

	return nil
}

// OutputList prints interfaces in list format
func OutputList(interfaces []*Interface, opts *Options) error {
	if opts.JSONOutput {
		return outputJSON(interfaces)
	}

	cols := NormalizeColumns(opts.Columns)

	// Print headers
	if !opts.NoHeadings {
		printHeaders(cols)
	}

	// Print each interface
	for _, iface := range interfaces {
		printInterfaceRow(iface, "", cols, opts)
	}

	return nil
}

// printHeaders prints column headers
func printHeaders(cols []string) {
	headers := make([]string, len(cols))
	for i, col := range cols {
		if colDef, ok := AvailableColumns[col]; ok {
			headers[i] = padRight(colDef.Name, colDef.Width)
		} else {
			headers[i] = padRight(col, 12)
		}
	}
	fmt.Println(strings.Join(headers, " "))
}

// printInterface prints an interface and its children in tree format
func printInterface(iface *Interface, prefix string, isLast bool, cols []string, opts *Options) {
	// Print this interface
	printInterfaceRow(iface, prefix, cols, opts)

	// Print children
	childCount := len(iface.Children)
	for i, child := range iface.Children {
		isLastChild := i == childCount-1

		// Build prefix for child
		var childPrefix string
		if prefix == "" {
			if isLastChild {
				childPrefix = "└─"
			} else {
				childPrefix = "├─"
			}
		} else {
			if isLast {
				childPrefix = prefix + "   "
			} else {
				childPrefix = prefix + "│  "
			}
			if isLastChild {
				childPrefix += "└─"
			} else {
				childPrefix += "├─"
			}
		}

		printInterface(child, childPrefix, isLastChild, cols, opts)
	}
}

// printInterfaceRow prints a single interface row
func printInterfaceRow(iface *Interface, treePrefix string, cols []string, opts *Options) {
	values := make([]string, len(cols))

	for i, col := range cols {
		width := AvailableColumns[col].Width
		value := getColumnValue(iface, col, opts)

		// For NAME column, prepend tree prefix
		if col == "NAME" && treePrefix != "" {
			value = treePrefix + value
			// Adjust width for tree characters
			width += len(treePrefix)
		}

		values[i] = padRight(value, width)
	}

	fmt.Println(strings.Join(values, " "))
}

// getColumnValue returns the value for a specific column
func getColumnValue(iface *Interface, col string, opts *Options) string {
	switch col {
	case "NAME":
		return iface.Name

	case "TYPE":
		return iface.Type

	case "STATE":
		return iface.State

	case "IP":
		return formatIPs(iface.IPs, opts.ShowAllIPs)

	case "IPV4":
		return formatIPList(iface.IPv4Addrs, opts.ShowAllIPs)

	case "IPV6":
		return formatIPList(iface.IPv6Addrs, opts.ShowAllIPs)

	case "MAC":
		if iface.MAC != nil {
			return iface.MAC.String()
		}
		return "-"

	case "MTU":
		if iface.MTU > 0 {
			return fmt.Sprintf("%d", iface.MTU)
		}
		return "-"

	case "DRIVER":
		if iface.Driver != "" {
			return iface.Driver
		}
		return "-"

	case "MODEL":
		if iface.Model != "" {
			return iface.Model
		}
		return "-"

	case "SPEED":
		if iface.Speed > 0 {
			return fmt.Sprintf("%dMb/s", iface.Speed)
		}
		return "-"

	case "MASTER":
		if iface.Master != "" {
			return iface.Master
		}
		return "-"

	case "PEER":
		if iface.Peer != "" {
			if iface.PeerNS != "" && iface.PeerNS != iface.Namespace {
				return fmt.Sprintf("%s@%s", iface.Peer, iface.PeerNS)
			}
			return iface.Peer
		}
		return "-"

	case "NAMESPACE":
		return iface.Namespace

	case "RX":
		if iface.RxBytes > 0 {
			return formatBytes(iface.RxBytes)
		}
		return "-"

	case "TX":
		if iface.TxBytes > 0 {
			return formatBytes(iface.TxBytes)
		}
		return "-"

	default:
		return "-"
	}
}

// formatIPs formats IP addresses for display
func formatIPs(ips []net.IP, showAll bool) string {
	if len(ips) == 0 {
		return "-"
	}

	if showAll {
		strs := make([]string, len(ips))
		for i, ip := range ips {
			strs[i] = ip.String()
		}
		return strings.Join(strs, ", ")
	}

	// Show primary + count
	primary := ips[0].String()
	if len(ips) > 1 {
		return fmt.Sprintf("%s (+%d)", primary, len(ips)-1)
	}
	return primary
}

// formatIPList formats a list of IPs
func formatIPList(ips []net.IP, showAll bool) string {
	if len(ips) == 0 {
		return "-"
	}

	if showAll {
		strs := make([]string, len(ips))
		for i, ip := range ips {
			strs[i] = ip.String()
		}
		return strings.Join(strs, ", ")
	}

	primary := ips[0].String()
	if len(ips) > 1 {
		return fmt.Sprintf("%s (+%d)", primary, len(ips)-1)
	}
	return primary
}

// formatBytes formats bytes in human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// padRight pads a string to the right
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// outputJSON outputs interfaces as JSON
func outputJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
