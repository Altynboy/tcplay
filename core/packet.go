package core

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"syscall"
	"time"
)

func (c *TCPConnection) sendPacket(header *TCPHeader) error {
	header.Checksum = c.calculateChecksum(header, nil, c.srcIP, c.destIP)

	log.Printf("Sending packet: %+v", header)

	buf := make([]byte, 20)
	binary.BigEndian.PutUint16(buf[0:2], header.SourcePort)
	binary.BigEndian.PutUint16(buf[2:4], header.DestPort)
	binary.BigEndian.PutUint32(buf[4:8], header.SeqNum)
	binary.BigEndian.PutUint32(buf[8:12], header.AckNum)
	buf[12] = header.HeaderLen << 4
	buf[13] = header.ControlFlags
	binary.BigEndian.PutUint16(buf[14:16], header.WindowSize)
	binary.BigEndian.PutUint16(buf[16:18], header.Checksum)
	binary.BigEndian.PutUint16(buf[18:20], header.UrgentPtr)

	addr := &syscall.SockaddrInet4{
		Port: int(c.destPort),
		Addr: c.destIP,
	}

	// mssOption := []byte{0x02, 0x04, 0x05, 0xb4}
	// buf = append(buf, mssOption...)

	if err := syscall.Sendto(c.rawSocket, buf, 0, addr); err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (c *TCPConnection) receivePacket() (*TCPHeader, error) {
	buf := make([]byte, 65535)
	for {
		log.Println("Receiving package")
		n, _, err := syscall.Recvfrom(c.rawSocket, buf, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to receive packet: %v", err)
		}

		log.Printf("%d packet send", n)
		if n < 20 {
			continue
		}

		ipHeaderLen := int(buf[0]&0x0F) * 4
		if n < ipHeaderLen+20 {
			log.Println("packet less that ipHeaderLen")
			continue
		}

		ipProtocol := buf[9]
		if ipProtocol != 6 { // TCP protocol number
			log.Println("protocol not TCP")
			continue
		}

		tcpHeaderData := buf[ipHeaderLen : ipHeaderLen+20]

		tcpHeader := &TCPHeader{
			SourcePort:   binary.BigEndian.Uint16(tcpHeaderData[0:2]),
			DestPort:     binary.BigEndian.Uint16(tcpHeaderData[2:4]),
			SeqNum:       binary.BigEndian.Uint32(tcpHeaderData[4:8]),
			AckNum:       binary.BigEndian.Uint32(tcpHeaderData[8:12]),
			HeaderLen:    tcpHeaderData[12] >> 4,
			ControlFlags: tcpHeaderData[13] & 0x3F,
			WindowSize:   binary.BigEndian.Uint16(tcpHeaderData[14:16]),
			Checksum:     binary.BigEndian.Uint16(tcpHeaderData[16:18]),
			UrgentPtr:    binary.BigEndian.Uint16(tcpHeaderData[18:20]),
		}

		log.Printf("Sending packet: %+v\n", tcpHeader)

		if tcpHeader.SourcePort == c.destPort && tcpHeader.DestPort == c.srcPort {
			return tcpHeader, nil
		}
	}
}

func (c *TCPConnection) waitForACK() error {
	return nil
}

func (c *TCPConnection) waitForFIN() error {
	return nil
}

func (c *TCPConnection) sendPacketWithPayload(header *TCPHeader, payload []byte) error {
	headerBytes := header.Serialize()

	packet := append(headerBytes, payload...)

	addr := &syscall.SockaddrInet4{
		Port: int(c.destPort),
		Addr: c.destIP,
	}

	if err := syscall.Sendto(c.rawSocket, packet, 0, addr); err != nil {
		return fmt.Errorf("failed to send packet with payload: %v", err)
	}

	return nil
}

// func (c *TCPConnection) calculateChecksum(header *TCPHeader) uint16 {

// 	return uint16(0)
// }

func (h *TCPHeader) Serialize() []byte {
	header := make([]byte, 20) // Minimum TCP header size

	// Fill in the fields
	binary.BigEndian.PutUint16(header[0:2], h.SourcePort)
	binary.BigEndian.PutUint16(header[2:4], h.DestPort)
	binary.BigEndian.PutUint32(header[4:8], h.SeqNum)
	binary.BigEndian.PutUint32(header[8:12], h.AckNum)

	// Data offset (5 for no options) and flags
	header[12] = (5 << 4) // Data offset (5 * 4 = 20 bytes header)
	header[13] = byte(h.ControlFlags)

	binary.BigEndian.PutUint16(header[14:16], h.WindowSize)
	binary.BigEndian.PutUint16(header[16:18], 0) // Checksum placeholder
	binary.BigEndian.PutUint16(header[18:20], h.UrgentPtr)

	// Compute checksum (pseudo-header + TCP header + payload)
	// checksum := computeChecksum(header)
	// binary.BigEndian.PutUint16(header[16:18], checksum)

	return header
}

func generateRandomSeqNum() uint32 {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rng.Uint32()
}
