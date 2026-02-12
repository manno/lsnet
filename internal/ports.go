package internal

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// ListeningPort represents a listening port on a specific address
type ListeningPort struct {
	Protocol string // "tcp" or "udp"
	Address  net.IP
	Port     uint16
}

// DiscoverListeningPorts discovers all listening TCP and UDP ports
func DiscoverListeningPorts() ([]ListeningPort, error) {
	var ports []ListeningPort

	// Parse TCP listening ports
	tcpPorts, err := parseNetFile("/proc/net/tcp", "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to parse TCP ports: %w", err)
	}
	ports = append(ports, tcpPorts...)

	// Parse TCP6 listening ports
	tcp6Ports, err := parseNetFile("/proc/net/tcp6", "tcp")
	if err != nil {
		// TCP6 may not exist on systems without IPv6, don't treat as fatal
		// Continue without error
	} else {
		ports = append(ports, tcp6Ports...)
	}

	// Parse UDP listening ports
	udpPorts, err := parseNetFile("/proc/net/udp", "udp")
	if err != nil {
		return nil, fmt.Errorf("failed to parse UDP ports: %w", err)
	}
	ports = append(ports, udpPorts...)

	// Parse UDP6 listening ports
	udp6Ports, err := parseNetFile("/proc/net/udp6", "udp")
	if err != nil {
		// UDP6 may not exist on systems without IPv6, don't treat as fatal
	} else {
		ports = append(ports, udp6Ports...)
	}

	return ports, nil
}

// parseNetFile parses /proc/net/tcp, /proc/net/tcp6, /proc/net/udp, or /proc/net/udp6
func parseNetFile(filename, protocol string) ([]ListeningPort, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []ListeningPort
	scanner := bufio.NewScanner(file)

	// Skip header line
	if !scanner.Scan() {
		return ports, nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		// Need at least 4 fields: sl, local_address, rem_address, st
		if len(fields) < 4 {
			continue
		}

		// Field 1 is local_address (address:port in hex)
		localAddr := fields[1]

		// Field 3 is st (state)
		// For TCP: 0A = LISTEN (10 in decimal)
		// For UDP: 07 = CLOSE (listening state)
		state := fields[3]

		// Only process listening sockets
		isListening := false
		if protocol == "tcp" && state == "0A" {
			isListening = true
		} else if protocol == "udp" && state == "07" {
			isListening = true
		}

		if !isListening {
			continue
		}

		// Parse local address
		ip, port, err := parseSocketAddr(localAddr)
		if err != nil {
			continue
		}

		ports = append(ports, ListeningPort{
			Protocol: protocol,
			Address:  ip,
			Port:     port,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ports, nil
}

// parseSocketAddr parses a socket address in the format "01020304:0050" or IPv6 format
func parseSocketAddr(addr string) (net.IP, uint16, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid address format: %s", addr)
	}

	// Parse IP address
	ipHex := parts[0]
	var ip net.IP

	if len(ipHex) == 8 {
		// IPv4 address (4 bytes = 8 hex chars)
		ipBytes, err := hex.DecodeString(ipHex)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decode IPv4 address: %w", err)
		}
		// The bytes are in little-endian order, reverse them
		ip = net.IPv4(ipBytes[3], ipBytes[2], ipBytes[1], ipBytes[0])
	} else if len(ipHex) == 32 {
		// IPv6 address (16 bytes = 32 hex chars)
		ipBytes, err := hex.DecodeString(ipHex)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decode IPv6 address: %w", err)
		}
		// IPv6 bytes are stored in groups of 4 bytes in little-endian order
		// Reverse each group of 4 bytes
		for i := 0; i < 16; i += 4 {
			ipBytes[i], ipBytes[i+1], ipBytes[i+2], ipBytes[i+3] =
				ipBytes[i+3], ipBytes[i+2], ipBytes[i+1], ipBytes[i]
		}
		ip = net.IP(ipBytes)
	} else {
		return nil, 0, fmt.Errorf("invalid IP address length: %d", len(ipHex))
	}

	// Parse port (in hex)
	portNum, err := strconv.ParseUint(parts[1], 16, 16)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse port: %w", err)
	}

	return ip, uint16(portNum), nil
}

// GetPortsForIP returns all listening ports for a specific IP address
// Note: Wildcard addresses (0.0.0.0 or ::) are excluded
func GetPortsForIP(ports []ListeningPort, ip net.IP) []ListeningPort {
	var result []ListeningPort

	for _, p := range ports {
		// Only include ports bound to this specific IP (not wildcards)
		if p.Address.Equal(ip) {
			result = append(result, p)
		}
	}

	return result
}

// isWildcardAddr returns true if the IP is a wildcard address (0.0.0.0 or ::)
func isWildcardAddr(ip net.IP) bool {
	return ip.IsUnspecified()
}
