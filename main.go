package main

import (
	"flag"
	"fmt"
	"os"

	wakeonlan "github.com/johndo100/wakeonlan/pkg/magic"
)

func main() {
	// Define command-line flags
	macAddr := flag.String("mac", "", "Target MAC address (required). Format: XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX")
	ip := flag.String("ip", "255.255.255.255", "Broadcast IP address (default: 255.255.255.255)")
	port := flag.String("port", "9", "Destination port: 0 (any), 7 (echo), or 9 (discard, default)")
	helpFlag := flag.Bool("help", false, "Show this help message")

	flag.Parse()

	// Show help if requested
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Validate required MAC address
	if *macAddr == "" {
		fmt.Fprintf(os.Stderr, "Error: MAC address is required\n\n")
		printHelp()
		os.Exit(1)
	}

	// Send magic packet
	fmt.Printf("Sending Wake-on-LAN magic packet...\n")
	fmt.Printf("  MAC Address: %s\n", *macAddr)
	fmt.Printf("  Broadcast IP: %s\n", *ip)
	fmt.Printf("  Port: %s\n", *port)
	fmt.Println()

	err := wakeonlan.SendMagic(*macAddr, "", *ip, *port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Magic packet sent successfully")
}

func printHelp() {
	fmt.Fprintf(os.Stderr, `Wake-on-LAN Magic Packet Sender

Usage:
  wakeonlan -mac <address> [options]

Examples:
  # Send to a specific MAC address using default broadcast
  wakeonlan -mac 00:11:22:33:44:55

  # Send to a specific IP and port
  wakeonlan -mac 00:11:22:33:44:55 -ip 192.168.1.255 -port 9

  # Use dash-separated MAC format
  wakeonlan -mac 00-11-22-33-44-55 -ip 192.168.1.100

Options:
`)
	flag.PrintDefaults()
}
