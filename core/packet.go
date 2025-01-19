package core

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"syscall"
	"tcplay/core/ip"
	"tcplay/protocol"
	"time"
)

func (c *TCPConnection) sendPacket(header *protocol.TCPHeader) error {
	// c.ipHeader.TotalLen = uint16(40)
	// ipHeader := c.ipHeader.Marshall()
	// fmt.Printf("ip header %v\n", ipHeader)
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

	// log.Printf("addr: %+v", addr)
	// log.Printf("socket %d", c.rawSocket)
	// mssOption := []byte{0x02, 0x04, 0x05, 0xb4}
	// buf = append(buf, mssOption...)

	// ipHeader = append(ipHeader, buf...)
	// log.Printf("packet len is %d", len(ipHeader))
	if err := syscall.Sendto(c.rawSocket, buf, 0, addr); err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (c *TCPConnection) ReceiveIPPacket() (*protocol.TCPHeader, error) {
	tcpHeaderData := ip.ReceivePacket(c.rawSocket)

	tcpHeader := &protocol.TCPHeader{
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

	if tcpHeader.SourcePort == c.destPort && tcpHeader.DestPort == c.srcPort {
		log.Printf("Received packet: %+v\n", tcpHeader)
		return tcpHeader, nil
	}

	return nil, fmt.Errorf("no packet send")
}

func (c *TCPConnection) ReceivePacket() (*protocol.TCPHeader, error) {
	buf := make([]byte, 65535)
	log.Println("Start receiving packets")
	for {
		n, _, err := syscall.Recvfrom(c.rawSocket, buf, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to receive packet: %v", err)
		}

		if n < 20 {
			continue
		}

		ipHeaderLen := int(buf[0]&0x0F) * 4
		if c.state == SYN_SENT {
			if n < ipHeaderLen+20 {
				log.Println("Skip packet len is less than < ipheaderLen + 20")
				continue
			}
		}

		ipProtocol := buf[9]
		if ipProtocol != 6 { // TCP protocol number
			log.Println("Skip packet, protocol is not TCP")
			log.Printf("Received packet protocol number: %v", buf[9])
			log.Printf("Received packet: %v", buf)

			continue
		}

		tcpHeaderData := buf[ipHeaderLen : ipHeaderLen+20]

		tcpHeader := &protocol.TCPHeader{
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

		if tcpHeader.SourcePort == c.destPort && tcpHeader.DestPort == c.srcPort {
			log.Printf("Received packet: %+v\n", tcpHeader)
			return tcpHeader, nil
		}
	}
}

func (c *TCPConnection) sendPacketWithPayload(header *protocol.TCPHeader, payload []byte) error {
	log.Printf("Sending packet with payload:\n %+v", header)

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

// func (c *TCPConnection) calculateChecksum(header *protocol.TCPHeader) uint16 {

// 	return uint16(0)
// }

func generateRandomSeqNum() uint32 {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rng.Uint32()
}
