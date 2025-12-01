package internal

import (
	"fmt"
	"strings"
)

// Column represents a displayable column
type Column struct {
	Name        string
	Description string
	Width       int
}

var (
	// DefaultColumns are the columns shown by default
	DefaultColumns = []string{"NAME", "TYPE", "STATE"}

	// AvailableColumns defines all columns that can be displayed
	AvailableColumns = map[string]Column{
		"NAME":      {"NAME", "Device/IP name", 20},
		"TYPE":      {"TYPE", "Type (ether, bridge, vlan, inet, inet6, etc.)", 10},
		"STATE":     {"STATE", "Interface state (UP/DOWN)", 6},
		"IP":        {"IP", "IP address summary (count)", 16},
		"IPV4":      {"IPV4", "IPv4 address count", 8},
		"IPV6":      {"IPV6", "IPv6 address count", 8},
		"MAC":       {"MAC", "MAC address", 17},
		"MTU":       {"MTU", "Maximum transmission unit", 6},
		"DRIVER":    {"DRIVER", "Kernel module/driver name", 12},
		"MODEL":     {"MODEL", "Device model/description", 20},
		"SPEED":     {"SPEED", "Link speed (e.g., 1000Mb/s)", 10},
		"MASTER":    {"MASTER", "Master device name (for bridge/bond members)", 12},
		"PEER":      {"PEER", "veth peer name (with namespace if different)", 20},
		"NAMESPACE": {"NAMESPACE", "Network namespace name", 12},
		"RX":        {"RX", "Received bytes/packets", 12},
		"TX":        {"TX", "Transmitted bytes/packets", 12},
	}
)

// ValidateColumns checks if all requested columns exist
func ValidateColumns(cols []string) error {
	for _, col := range cols {
		colUpper := strings.ToUpper(col)
		if _, exists := AvailableColumns[colUpper]; !exists {
			return fmt.Errorf("unknown column: %s (use --list-columns to see available columns)", col)
		}
	}
	return nil
}

// PrintAvailableColumns prints all available columns with descriptions
func PrintAvailableColumns() {
	fmt.Println("Available columns:")
	fmt.Println()

	// Print in a specific order for readability
	order := []string{
		"NAME", "TYPE", "STATE", "IP", "IPV4", "IPV6", "MAC", "MTU",
		"DRIVER", "MODEL", "SPEED", "MASTER", "PEER", "NAMESPACE", "RX", "TX",
	}

	for _, name := range order {
		if col, ok := AvailableColumns[name]; ok {
			fmt.Printf("  %-12s %s\n", col.Name, col.Description)
		}
	}

	fmt.Println()
	fmt.Printf("Default columns: %s\n", strings.Join(DefaultColumns, ","))
}

// NormalizeColumns converts column names to uppercase
func NormalizeColumns(cols []string) []string {
	normalized := make([]string, len(cols))
	for i, col := range cols {
		normalized[i] = strings.ToUpper(col)
	}
	return normalized
}
