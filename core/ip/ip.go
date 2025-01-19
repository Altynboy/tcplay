package ip

import (
	"encoding/binary"
	"fmt"
	"log"
	"syscall"
)

// IPv4 Header Format
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |Version|  IHL  |Type of Service|          Total Length           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |         Identification        |Flags|      Fragment Offset      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |  Time to Live |    Protocol   |         Header Checksum         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                       Source IP Address                         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                    Destination IP Address                       |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                    Options                    |    Padding      |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type IPHeader struct {
	Version    uint8
	IHL        uint8
	TOS        uint8
	TotalLen   uint16
	ID         uint16
	Flags      uint16
	FragOffset uint16
	TTL        uint8
	Protocol   uint8
	Checksum   uint16
	SrcAddr    [4]byte
	DstAddr    [4]byte
}

func (header *IPHeader) Marshall() []byte {
	// Convert header to bytes
	headerBytes := make([]byte, 20)
	// Version and IHL (first byte)
	headerBytes[0] = (header.Version << 4) | header.IHL
	headerBytes[1] = header.TOS
	binary.BigEndian.PutUint16(headerBytes[2:], header.TotalLen)
	binary.BigEndian.PutUint16(headerBytes[4:], header.ID)
	binary.BigEndian.PutUint16(headerBytes[6:], header.Flags<<13|header.FragOffset)
	headerBytes[8] = header.TTL
	headerBytes[9] = header.Protocol
	copy(headerBytes[12:16], header.SrcAddr[:])
	copy(headerBytes[16:20], header.DstAddr[:])

	binary.BigEndian.PutUint16(headerBytes[10:], 0)
	header.Checksum = CalculateChecksum(headerBytes)
	binary.BigEndian.PutUint16(headerBytes[10:], header.Checksum)

	log.Printf("IP Header packet: %+v\n", header)
	return headerBytes
}

func Serialize(data []byte) (*IPHeader, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("packet too short for IP header: %d bytes", len(data))
	}

	header := &IPHeader{
		Version:    uint8(data[0] >> 4),
		IHL:        uint8(data[0] & 0x0F),
		TOS:        uint8(data[1]),
		TotalLen:   binary.BigEndian.Uint16(data[2:4]),
		ID:         binary.BigEndian.Uint16(data[4:6]),
		FragOffset: binary.BigEndian.Uint16(data[6:8]),
		TTL:        uint8(data[8]),
		Protocol:   uint8(data[9]),
		Checksum:   binary.BigEndian.Uint16(data[10:12]),
		SrcAddr:    [4]byte(data[12:16]),
		DstAddr:    [4]byte(data[16:20]),
	}

	return header, nil
}

func ReceivePacket(fd int) []byte {
	// Buffer for receiving data
	buf := make([]byte, 65536)

	for {
		// Read from the socket
		n, addr, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			fmt.Printf("Error receiving packet: %v\n", err)
			continue
		}

		// Parse IP header
		ipHeader, err := Serialize(buf[:n])
		if err != nil {
			fmt.Printf("Error parsing IP header: %v\n", err)
			continue
		}

		// Print packet information
		fmt.Printf("\n--- New Packet ---\n")
		fmt.Printf("From: %v\n", ipHeader.SrcAddr)
		fmt.Printf("To: %v\n", ipHeader.DstAddr)
		fmt.Printf("Protocol: %d\n", ipHeader.Protocol)
		fmt.Printf("TTL: %d\n", ipHeader.TTL)
		fmt.Printf("Length: %d\n", ipHeader.TotalLen)
		fmt.Printf("ID: %d\n", ipHeader.ID)

		// If this is a TCP packet (protocol 6), you can parse the TCP header here
		if ipHeader.Protocol == 6 {
			// The TCP header starts after the IP header
			tcpStart := int(ipHeader.IHL) * 4
			if n > tcpStart {
				tcpData := buf[tcpStart:n]
				// Parse TCP data here if needed
				srcPort := binary.BigEndian.Uint16(tcpData[0:2])
				dstPort := binary.BigEndian.Uint16(tcpData[2:4])
				fmt.Printf("Source Port: %d\n", srcPort)
				fmt.Printf("Destination Port: %d\n", dstPort)
				return tcpData
			}
		}

		fmt.Printf("Remote address: %v\n", addr)
	}
}

func CalculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)
	return ^uint16(sum)
}
