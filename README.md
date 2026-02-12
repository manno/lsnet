# lsnet - Network Device Tree Viewer

A command-line tool to display network devices and their relationships in a tree format, similar to `lsblk` for block devices.

## Features

- **Tree view** of network interfaces showing hierarchies (bridges, VLANs, veth pairs, etc.)
- **IP addresses as tree nodes** with CIDR notation
- **Listening ports** shown under each IP address (TCP and UDP)
- **Flexible column display** like `lsblk` with `-o` flag
- **Network namespace** aware (future enhancement)
- **Driver and hardware** information from sysfs
- **JSON output** for scripting

## Installation

```bash
go build -o lsnet
sudo install lsnet /usr/local/bin/
```

## Usage

### Basic Examples

```bash
# Show active interfaces with default columns
lsnet

# Show all interfaces including DOWN
lsnet -a

# Custom columns
lsnet -o NAME,TYPE,DRIVER,IP

# Append columns to defaults
lsnet -o+MAC,MTU,DRIVER

# List all available columns
lsnet --list-columns

# JSON output for scripting
lsnet -J
```

### Example Output

```
NAME                 TYPE       STATE
eth0                 ether      UP
├─192.168.1.10/24      inet
│ ├─22                 tcp
│ ├─80                 tcp
│ └─443                tcp
├─vlan100            vlan       UP
│ └─10.0.1.1/24        inet
│   └─8080             tcp
└─vlan200            vlan       UP
  └─10.0.2.1/24        inet
br0                  bridge     UP
├─172.16.0.1/16        inet
│ └─53                 udp
├─fe80::1/64           inet6
├─eth1               ether      UP
└─veth0              veth       UP
wlan0                ether      UP
└─192.168.0.50/24      inet
```

## Available Columns

| Column | Description |
|--------|-------------|
| NAME | Device/IP name (includes CIDR for IP addresses) |
| TYPE | Type (ether, bridge, vlan, inet, inet6, etc.) |
| STATE | Interface state (UP/DOWN, empty for IP addresses) |
| IP | IP address count (total IPv4 + IPv6) |
| IPV4 | IPv4 address count |
| IPV6 | IPv6 address count |
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

Default columns: `NAME,TYPE,STATE`

**Note:** IP addresses are displayed as child nodes in the tree with CIDR notation, not as columns. Listening ports (TCP and UDP) are shown as child nodes under their corresponding IP addresses. Only ports bound to specific IP addresses are shown; wildcard addresses (`0.0.0.0` or `::`) are excluded to keep the output concise.

## Command-Line Options

```
  -a, --all              show all interfaces including DOWN
  -N, --all-namespaces   show interfaces from all network namespaces
  -o, --output <list>    output columns (comma-separated, or +COL to append)
  --list-columns         list all available columns
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
