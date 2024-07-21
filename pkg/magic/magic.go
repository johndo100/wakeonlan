package wakeonlan

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

// The magic packet is a frame that contains anywhere within its payload
// for a total of 102 bytes. (6 bytes + 16 * 48-bits)
// https://www.ti.com/lit/an/snla261a/snla261a.pdf

/*
 *	How a Magic Packet Frame looks like:
 *
 *	 _________________________________________
 *	| 0xff | 0xff | 0xff | 0xff | 0xff | 0xff |
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	|MAC[0]|MAC[1]|MAC[2]|MAC[3]|MAC[4]|MAC[5]|
 *	 -----------------------------------------
 *	 optional SecureON (tm) password:
 *	 _________________________________________
 *	|PASS0 |PASS1 |PASS2 |PASS3 |PASS4 |PASS5 |
 *	 -----------------------------------------
 */

type Packet struct {
	payload bytes.Buffer
}

// 6 bytes of all 255 (FF FF FF FF FF FF in hexadecimal, 11111111 in binary)
func (p *Packet) writeHeader() {
	for range 6 {
		ff := []byte{0b11111111}
		fmt.Print(ff)
		p.payload.Write(ff)
	}
}

// followed by 16 repetitions of the target computer's 48-bit MAC address
func (p *Packet) WriteMAC(addr string) error {
	// parses an IEEE 802 MAC-48, EUI-48, EUI-64, or a 20-octet
	// check mac address by built-in lib: https://go.dev/src/net/mac.go
	// accept only EUI-48: 17 or 14 chars string
	hwAddr, err := net.ParseMAC(addr)

	switch {
	case err != nil:
		return err
	case !(len(addr) == 17 || len(addr) == 14):
		err := fmt.Errorf("address %s: invalid EUI-48 format", addr)
		return err
	default:
		for range 16 {
			p.payload.Write(hwAddr)
		}
		return nil
	}
}

// optional SecureON (tm) password
// implement passwd method to append
// more [6]byte in the end of the payload
// func (p *Packet) passwd(p string) {

// }

func (p *Packet) SendUDP(ip, port string) error {
	// SUPPORT IPV4 ONLY
	const network = "udp4"
	// set port first
	remoteIP, err := getRAddr(ip)
	if err != nil {
		fmt.Println(err)
	}
	remotePort, err := getPort(port)
	if err != nil {
		fmt.Println(err)
	}

	raddr := net.UDPAddr{
		IP:   remoteIP,
		Port: remotePort,
	}

	conn, err := net.DialUDP(network, nil, &raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Printf("The payload: %X.\n", p.payload.Bytes())
	return nil
}

// remote ip address
func getRAddr(addr string) (net.IP, error) {
	dfltRAddr := net.IPv4(255, 255, 255, 255) // Default: send to broadcast address of the local network
	rAddr := net.ParseIP(addr)
	if rAddr == nil {
		return dfltRAddr, fmt.Errorf("remote address %v: invalid, but we still try to send magic over your local network", addr)
	}
	return rAddr, nil
}

// port to send
func getPort(port string) (int, error) {
	acptPort := [3]int{0, 7, 9} // port 0, 7 (Echo Protocol) or 9 (Discard Protocol)
	dfltPort := 9
	ps, err := strconv.Atoi(port)
	if err != nil {
		ps = dfltPort
		fmt.Print(ps)
	}
	for _, v := range acptPort {
		if ps == v {
			return v, nil
		}
	}
	return dfltPort, fmt.Errorf("port %d: invalid, we've selected port %d for you", ps, dfltPort)
}

func SendMagic(macAddr, passwd, ip, port string) error {
	pk := new(Packet)

	// without password
	if passwd == "" {
		// assemble magic header
		pk.writeHeader()
		// and now the mac address
		err := pk.WriteMAC(macAddr)
		if err != nil {
			fmt.Print(err)
		} else {
			fmt.Println("Send magic packet without password...")
		}
	} else {
		// with password
		fmt.Println("We don't support Wake-on-LAN password yet...")
	}
	// then send via UDP4
	err := pk.SendUDP(ip, port)
	if err != nil {
		return err
	}
	return nil
}
