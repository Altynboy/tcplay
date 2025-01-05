package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
)

func (c *TCPConnection) sendPacket(header *TCPHeader) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, header); err != nil {
		return fmt.Errorf("failed to prepare packet to send: %v", err)
	}

	addr := &syscall.SockaddrInet4{
		Port: int(c.destPort),
		Addr: c.destIP,
	}

	if err := syscall.Sendmsg(c.rawSocket, buf.Bytes(), nil, addr, 0); err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (c *TCPConnection) receivePacket() (*TCPHeader, error) {
	buf := make([]byte, 65535)
	for {
		n, _, err := syscall.Recvfrom(c.rawSocket, buf, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to receive packet: %v", err)
		}

		if n < 20 {
			continue
		}

		ipHeaderLen := int(buf[0]&0x0F) * 4
		if n < ipHeaderLen+20 {
			continue
		}

		tcpHeaderData := buf[ipHeaderLen : ipHeaderLen+20]
		// TODO: rewrite using BigEndian
		tcpHeader := &TCPHeader{
			SourcePort:   uint16(tcpHeaderData[0])<<8 | uint16(tcpHeaderData[1]),
			DestPort:     uint16(tcpHeaderData[2])<<8 | uint16(tcpHeaderData[3]),
			SeqNum:       uint32(tcpHeaderData[4])<<24 | uint32(tcpHeaderData[5])<<16 | uint32(tcpHeaderData[6])<<8 | uint32(tcpHeaderData[7]),
			AckNum:       uint32(tcpHeaderData[8])<<24 | uint32(tcpHeaderData[9])<<16 | uint32(tcpHeaderData[10])<<8 | uint32(tcpHeaderData[11]),
			ControlFlags: tcpHeaderData[13] & 0x3F,
			WindowSize:   uint16(tcpHeaderData[14]) << 8,
		}

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

func (c *TCPConnection) calculateChecksum(header *TCPHeader) uint16 {

	return uint16(0)
}
