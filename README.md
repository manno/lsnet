# netree - Network Device Tree Viewer

A command-line tool to display network devices and their relationships in a tree format, similar to `lsblk` for block devices.

## Features

- **Tree view** of network interfaces showing hierarchies (bridges, VLANs, veth pairs, etc.)
- **Flexible column display** like `lsblk` with `-o` flag
- **Multiple IP addresses** support with smart display
- **Network namespace** aware (future enhancement)
- **Driver and hardware** information from sysfs
- **JSON output** for scripting

## Installation

```bash
go build -o netree
sudo install netree /usr/local/bin/
```

## Usage

### Basic Examples

```bash
# Show active interfaces with default columns
netree

# Show all interfaces including DOWN
netree -a

# Custom columns
netree -o NAME,TYPE,DRIVER,IP

# Append columns to defaults
netree -o +MAC,MTU,DRIVER

# Show all IP addresses (not just primary)
netree --all-ips

# List all available columns
netree --list-columns

# JSON output for scripting
netree -J
```

### Example Output

```
NAME         TYPE      STATE  IP
eth0         ether     UP     192.168.1.10
├─vlan100    vlan      UP     10.0.1.1
└─vlan200    vlan      UP     10.0.2.1
br0          bridge    UP     172.16.0.1 (+2)
├─eth1       ether     UP     -
└─veth0      veth      UP     -
wlan0        ether     UP     192.168.0.50
```

## Available Columns

| Column | Description |
|--------|-------------|
| NAME | Device name |
| TYPE | Interface type (ether, bridge, vlan, veth, bond, etc.) |
| STATE | Interface state (UP/DOWN) |
| IP | IP address(es) - shows primary with count if multiple |
| IPV4 | IPv4 addresses only |
| IPV6 | IPv6 addresses only |
| MAC | MAC address |
| MTU | Maximum transmission unit |
| DRIVER | Kernel module/driver name |
| MODEL | Device model/description |
| SPEED | Link speed (e.g., 1000Mb/s) |
| MASTER | Master device name (for bridge/bond members) |
| PEER | veth peer name (with namespace if different) |
| NAMESPACE | Network namespace name |
| RX | Received bytes/packets |
| TX | Transmitted bytes/packets |

Default columns: `NAME,TYPE,STATE,IP`

## Command-Line Options

```
  -a, --all              show all interfaces including DOWN
  -N, --all-namespaces   show interfaces from all network namespaces
  -o, --output <list>    output columns (comma-separated, or +COL to append)
  --list-columns         list all available columns
  --all-ips              show all IP addresses (not just primary)
  -d                     tree direction: logical devices down
  -u                     tree direction: physical devices up (default)
  -l, --list             list format (no tree)
  -J, --json             JSON output
  -n, --noheadings       don't print column headers
  -t, --type <types>     show only specified types (comma-separated)
  -x, --exclude <types>  exclude specified types (comma-separated)
  -v, --version          show version
```

## Design

See [DESIGN.md](DESIGN.md) for detailed design documentation including:
- Architecture and data structures
- Hierarchy rules
- Network namespace handling
- Future enhancements

## Requirements

- Linux (uses netlink and sysfs)
- Root or CAP_NET_ADMIN capabilities for full functionality

## Dependencies

- [github.com/vishvananda/netlink](https://github.com/vishvananda/netlink) - netlink library
- [github.com/vishvananda/netns](https://github.com/vishvananda/netns) - network namespace handling

## License

MIT

## Similar Tools

- `lsblk` - list block devices in a tree
- `ip link` - show network interfaces (flat list)
- `bridge link` - show bridge members
- `nmcli device` - NetworkManager device list
