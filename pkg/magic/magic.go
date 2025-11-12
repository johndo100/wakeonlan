// Package wakeonlan provides functionality to send Wake-on-LAN (WoL) magic packets
// to wake up computers on a local network.
//
// Wake-on-LAN (WoL) is a networking standard that allows a computer to be
// powered on or awakened remotely by sending a special formatted network packet
// called a "magic packet" to the target device's MAC address.
//
// Basic usage:
//
//	err := SendMagic("00:11:22:33:44:55", "", "255.255.255.255", "9")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Magic Packet Format:
// The magic packet is a frame that contains 6 bytes of 0xFF followed by the target
// MAC address repeated 16 times, for a total of 102 bytes (6 + 16*6).
// Reference: https://www.ti.com/lit/an/snla261a/snla261a.pdf
//
// Example packet structure (6 bytes + 16 repetitions of MAC):
//
//	 _________________________________________
//	| 0xff | 0xff | 0xff | 0xff | 0xff | 0xff |
//	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
//	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
//	|           ... (16 total repetitions)     |
//	 -----------------------------------------
package wakeonlan

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

// Packet represents a Wake-on-LAN magic packet that can be constructed and sent.
// Use SendMagic as a convenience function, or construct and populate a Packet
// manually using WriteMAC followed by SendUDP for more control.
type Packet struct {
	payload bytes.Buffer
}

// writeHeader writes 6 bytes of 0xFF to the packet payload.
// These bytes are the magic packet header.
func (p *Packet) writeHeader() {
	for range 6 {
		p.payload.WriteByte(0xFF)
	}
}

// WriteMAC writes the target MAC address 16 times to the magic packet payload.
//
// The MAC address must be in IEEE 802 MAC-48 format, using either colon or
// dash separators (e.g., "00:11:22:33:44:55" or "00-11-22-33-44-55").
//
// Parameters:
//   - addr: Target MAC address to write to the payload.
//
// Returns:
//   - nil on success
//   - error if the MAC address is invalid or malformed
//
// Example:
//
//	packet := &Packet{}
//	packet.writeHeader() // called automatically by SendMagic
//	err := packet.WriteMAC("00:11:22:33:44:55")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Packet) WriteMAC(addr string) error {
	// Parse using built-in net.ParseMAC
	// Accepts formats: "01:23:45:67:89:ab" or "01-23-45-67-89-ab"
	hwAddr, err := net.ParseMAC(addr)
	if err != nil {
		return fmt.Errorf("invalid MAC address %q: %w", addr, err)
	}

	// Write the MAC address 16 times to the payload
	for range 16 {
		p.payload.Write(hwAddr)
	}
	return nil
}

// optional SecureON (tm) password
// implement passwd method to append
// more [6]byte in the end of the payload
// func (p *Packet) passwd(p string) {

// }

// SendUDP sends the constructed magic packet via UDP to the target address and port.
//
// This method transmits the magic packet using UDP/IPv4. The packet must have been
// populated with a MAC address using WriteMAC before calling this method.
//
// Parameters:
//   - ip: Target IP address (typically a broadcast address like 255.255.255.255).
//     If empty, uses the broadcast address by default.
//   - port: Destination port as a string. Accepted values: "0", "7", "9" (default).
//     Port 7: Echo Protocol, Port 9: Discard Protocol
//
// Returns:
//   - nil on successful transmission
//   - error if the packet cannot be sent (invalid IP, port, or network error)
//
// Example:
//
//	packet := &Packet{}
//	packet.writeHeader()
//	err := packet.WriteMAC("00:11:22:33:44:55")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	err = packet.SendUDP("255.255.255.255", "9")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Packet) SendUDP(ip, port string) error {
	// SUPPORT IPV4 ONLY
	const network = "udp4"

	remoteIP, err := getRAddr(ip)
	if err != nil {
		return fmt.Errorf("parse remote IP: %w", err)
	}

	remotePort, err := getPort(port)
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	raddr := net.UDPAddr{
		IP:   remoteIP,
		Port: remotePort,
	}

	conn, err := net.DialUDP(network, nil, &raddr)
	if err != nil {
		return fmt.Errorf("dial UDP: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(p.payload.Bytes())
	if err != nil {
		return fmt.Errorf("write UDP packet: %w", err)
	}

	return nil
}

// getRAddr parses the remote IP address.
// If addr is empty, returns the broadcast address (255.255.255.255).
// Returns an error if the address is invalid (non-empty and unparseable).
func getRAddr(addr string) (net.IP, error) {
	if addr == "" {
		// Empty address → use broadcast (default behavior)
		return net.IPv4(255, 255, 255, 255), nil
	}

	rAddr := net.ParseIP(addr)
	if rAddr == nil {
		return nil, fmt.Errorf("invalid remote address %q", addr)
	}
	return rAddr, nil
}

// getPort parses and validates the destination port.
// Accepted ports: 0 (any), 7 (echo), 9 (discard). Default is 9 if empty.
func getPort(port string) (int, error) {
	acceptedPorts := []int{0, 7, 9}
	defaultPort := 9

	// If empty, use default port
	if port == "" {
		return defaultPort, nil
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", port, err)
	}

	// Check if port is in accepted list
	for _, validPort := range acceptedPorts {
		if portNum == validPort {
			return portNum, nil
		}
	}

	// Port not in accepted list
	return 0, fmt.Errorf("port %d not supported (use 0, 7, or 9)", portNum)
}

// SendMagic sends a Wake-on-LAN magic packet to wake up a remote computer.
//
// This is the main API function that handles creating, populating, and sending
// a complete magic packet in a single call.
//
// Parameters:
//   - macAddr: Target MAC address of the computer to wake (required).
//     Format: "XX:XX:XX:XX:XX:XX" or "XX-XX-XX-XX-XX-XX"
//   - passwd: SecureON password (currently unsupported, pass empty string "").
//   - ip: Destination IP address, typically a broadcast address (e.g., "255.255.255.255").
//     If empty string is passed, defaults to broadcast address.
//   - port: Destination port as a string. Common values: "9" (discard), "7" (echo), "0" (any).
//     If empty string is passed, defaults to port 9.
//
// Returns:
//   - nil on successful packet transmission
//   - error describing what went wrong (invalid MAC, invalid IP, network error, etc.)
//
// Examples:
//
//	// Basic usage with defaults
//	err := SendMagic("00:11:22:33:44:55", "", "", "")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Specify target network broadcast
//	err := SendMagic("00:11:22:33:44:55", "", "192.168.1.255", "9")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Using dash-separated MAC format
//	err := SendMagic("00-11-22-33-44-55", "", "255.255.255.255", "9")
//	if err != nil {
//	    log.Fatal(err)
//	}
func SendMagic(macAddr, passwd, ip, port string) error {
	pk := new(Packet)

	// without password
	if passwd == "" {
		// assemble magic header
		pk.writeHeader()
		// write the MAC address 16 times
		err := pk.WriteMAC(macAddr)
		if err != nil {
			return fmt.Errorf("write MAC address: %w", err)
		}
	} else {
		// with password
		return fmt.Errorf("password-protected Wake-on-LAN not supported yet")
	}

	// send via UDP4
	err := pk.SendUDP(ip, port)
	if err != nil {
		return fmt.Errorf("send UDP: %w", err)
	}
	return nil
}
