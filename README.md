# wakeonlan

A lightweight Go library for sending Wake-on-LAN (WoL) magic packets to wake up computers on a local network.

[![Go Reference](https://pkg.go.dev/badge/github.com/johndo100/wakeonlan.svg)](https://pkg.go.dev/github.com/johndo100/wakeonlan)
[![Go Report Card](https://goreportcard.com/badge/github.com/johndo100/wakeonlan)](https://goreportcard.com/report/github.com/johndo100/wakeonlan)

## Overview

**Wake-on-LAN (WoL)** is a networking standard that allows a computer to be powered on or awakened remotely by sending a special formatted network packet (called a "magic packet") to the target device's MAC address.

This library provides:
- 🚀 Simple, easy-to-use API for sending magic packets
- 🔒 Proper error handling with context-aware error messages
- 🎯 Support for both broadcast and unicast addresses
- ⚡ High-performance packet construction (177 ns/op)
- 📦 Minimal dependencies (only Go standard library)
- 🧪 Comprehensive test suite (61.5% coverage, 16 tests)

## Installation

```bash
go get github.com/johndo100/wakeonlan
```

Requires Go 1.20 or later.

## Quick Start

### Basic Usage (Library)

```go
package main

import (
	"log"
	"github.com/johndo100/wakeonlan/pkg/magic"
)

func main() {
	// Send to a device using its MAC address
	err := magic.SendMagic("00:11:22:33:44:55", "", "", "")
	if err != nil {
		log.Fatal(err)
	}
}
```

### Command-Line Tool

The package includes a CLI tool for sending WoL packets:

```bash
# Build the CLI
go build

# Send to a specific MAC address (uses broadcast)
./wakeonlan -mac 00:11:22:33:44:55

# Send to specific network broadcast address
./wakeonlan -mac 00:11:22:33:44:55 -ip 192.168.1.255 -port 9

# Show help
./wakeonlan -help
```

## Usage Examples

### Example 1: Basic Wake-up with Defaults

```go
err := magic.SendMagic("00:11:22:33:44:55", "", "", "")
if err != nil {
    log.Fatal(err)
}
```

This sends a magic packet to the target MAC address using:
- Broadcast address: `255.255.255.255`
- Port: `9` (Discard Protocol)

### Example 2: Specific Network Broadcast

```go
err := magic.SendMagic("00:11:22:33:44:55", "", "192.168.1.255", "9")
if err != nil {
    log.Fatal(err)
}
```

Target the network broadcast address for more reliable delivery on specific subnets.

### Example 3: Alternative MAC Format

```go
// Both colon and dash separators are supported
err := magic.SendMagic("00-11-22-33-44-55", "", "255.255.255.255", "9")
if err != nil {
    log.Fatal(err)
}
```

### Example 4: Manual Packet Construction

```go
packet := &magic.Packet{}
packet.WriteMAC("00:11:22:33:44:55")
if err := packet.SendUDP("255.255.255.255", "9"); err != nil {
    log.Fatal(err)
}
```

## API Reference

### SendMagic (Main Function)

```go
func SendMagic(macAddr, passwd, ip, port string) error
```

Sends a Wake-on-LAN magic packet to wake up a remote computer.

**Parameters:**
- `macAddr`: Target MAC address (required). Format: `XX:XX:XX:XX:XX:XX` or `XX-XX-XX-XX-XX-XX`
- `passwd`: SecureON password (currently unsupported, pass empty string `""`)
- `ip`: Destination IP address. Empty string defaults to `255.255.255.255`
- `port`: Destination port as string. Supported: `"0"`, `"7"` (echo), `"9"` (discard, default). Empty defaults to `"9"`

**Returns:** `nil` on success, or an error describing what went wrong.

**Errors:**
- Invalid MAC address format
- Invalid IP address format
- Unsupported port number
- Network/UDP transmission errors

### Packet Type

```go
type Packet struct
```

Represents a Wake-on-LAN magic packet that can be constructed and sent manually.

**Methods:**

#### WriteMAC

```go
func (p *Packet) WriteMAC(addr string) error
```

Writes the target MAC address 16 times to the magic packet payload.

#### SendUDP

```go
func (p *Packet) SendUDP(ip, port string) error
```

Sends the constructed magic packet via UDP to the target address and port.

## Magic Packet Format

The magic packet structure is 102 bytes total:

- **Header (6 bytes):** `0xFF 0xFF 0xFF 0xFF 0xFF 0xFF`
- **Payload (96 bytes):** Target MAC address repeated 16 times

```
 _________________________________________
| 0xff | 0xff | 0xff | 0xff | 0xff | 0xff |  ← Header
|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
|                    ...                   |
|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|  ← 16 repetitions total
 -----------------------------------------
```

**Reference:** [Wake-on-LAN Technical Specification](https://www.ti.com/lit/an/snla261a/snla261a.pdf)

## Supported Ports

The library supports three standard ports for WoL:

| Port | Protocol | Use Case |
|------|----------|----------|
| `0` | Any | Send to any available port |
| `7` | Echo Protocol | Standard echo service |
| `9` | Discard Protocol | Standard discard service (default) |

## Supported MAC Address Formats

The library accepts MAC addresses in multiple formats:

```
00:11:22:33:44:55   (colon separated)
00-11-22-33-44-55   (dash separated)
aabbccddeeff        (no separator, lowercase)
AABBCCDDEEFF        (no separator, uppercase)
```

## Features

### Error Handling

The library provides clear, context-aware error messages:

```go
err := magic.SendMagic("invalid-mac", "", "255.255.255.255", "9")
if err != nil {
    // Prints: "write MAC address: invalid MAC address \"invalid-mac\": ..."
    log.Fatal(err)
}
```

### Network Support

- **IPv4 only** (UDP/IPv4)
- Supports both broadcast and unicast addressing
- Automatic defaults for optimal reliability

### Performance

Benchmarks (on Intel i7-8700):
- MAC address parsing: 177.9 ns/op
- IP address parsing: 56.81 ns/op
- Port parsing: 5.91 ns/op

## Testing

The package includes comprehensive tests with 61.5% code coverage:

```bash
# Run all tests
go test -v ./pkg/magic

# Run with coverage
go test -cover ./pkg/magic

# Run benchmarks
go test -bench=. ./pkg/magic -benchmem
```

**Test Coverage:**
- `WriteMAC`: 100% - MAC address parsing and validation
- `getRAddr`: 100% - IP address parsing
- `getPort`: 100% - Port validation
- `writeHeader`: 100% - Packet header construction
- `SendMagic`: 63.6% - Integration tests

## CLI Usage

A command-line tool is included for quick WoL operations:

```bash
# Build the CLI tool
go build

# Send WoL packet
./wakeonlan -mac 00:11:22:33:44:55

# Full options
./wakeonlan -mac <address> [-ip <broadcast>] [-port <port>]

# Show help
./wakeonlan -help
```

**Example Output:**
```
Sending Wake-on-LAN magic packet...
  MAC Address: 00:11:22:33:44:55
  Broadcast IP: 255.255.255.255
  Port: 9

✓ Magic packet sent successfully
```

## Limitations

- **Password-protected WoL:** SecureON passwords are not yet supported
- **IPv6:** Only IPv4 is supported
- **Network-dependent:** Success depends on network configuration and device support

## Requirements

- Go 1.20 or later
- UDP network access (required to send packets)
- Target device must have Wake-on-LAN enabled in BIOS

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See LICENSE file for details.

## References

- [Wake-on-LAN - Wikipedia](https://en.wikipedia.org/wiki/Wake-on-LAN)
- [WoL Technical Specification](https://www.ti.com/lit/an/snla261a/snla261a.pdf)
- [IEEE 802 MAC-48](https://en.wikipedia.org/wiki/MAC_address)

## Changelog

### v0.1.0 (Alpha)
- Initial release
- Basic magic packet construction and sending
- CLI tool with flag support
- Comprehensive error handling
- Full test suite with benchmarks