package core

import (
	"fmt"
	"log"
	"syscall"
	"tcplay/components/waiter"
	"tcplay/core/ip"
	"tcplay/protocol"
	"time"

	"math/rand"
)

type TCPConnection struct {
	srcPort    uint16
	destPort   uint16
	srcIP      [4]byte
	destIP     [4]byte
	seqNum     uint32
	ackNum     uint32
	state      uint8
	rawSocket  int
	receiveBuf []byte
	sendBuf    []byte
	maxSegSize uint16
	ipHeader   ip.IPHeader
}

const (
	CLOSED      = 0
	SYN_SENT    = 1
	ESTABLISHED = 2
)

func CreateConnection(destPort uint16, destIP [4]byte) (*TCPConnection, error) {
	sourcePort := uint16(49152 + rand.Intn(65535-49152+1))
	// srcIP := [4]byte{127, 0, 0, 1}
	srcIP := [4]byte{192, 168, 1, 103}

	// Create raw socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %v", err)
	}

	// err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	// if err != nil {
	// 	log.Printf("Setsockopt error: %v\n", err)
	// 	return nil, fmt.Errorf("failed to Setsockopt: %v", err)
	// }

	// Create IP header
	ipHeader := &ip.IPHeader{
		Version:  4,                   // IPv4
		IHL:      5,                   // 5 x 32-bit words
		TOS:      0,                   // Type of Service
		TotalLen: 40,                  // Header length
		ID:       1,                   // Identification
		TTL:      64,                  // Time to Live
		Protocol: syscall.IPPROTO_TCP, // TCP protocol
		SrcAddr:  srcIP,
		DstAddr:  srcIP,
	}

	// localAddr := &syscall.SockaddrInet4{
	// 	Port: int(sourcePort),
	// 	Addr: srcIP,
	// }

	// if err := syscall.Bind(fd, localAddr); err != nil {
	// 	return nil, fmt.Errorf("error binding client socket: %v", err)
	// }

	return &TCPConnection{
		srcPort:    sourcePort,
		destPort:   destPort,
		srcIP:      srcIP,
		destIP:     srcIP,
		seqNum:     generateRandomSeqNum(),
		ackNum:     0,
		state:      CLOSED,
		rawSocket:  fd,
		maxSegSize: 1460,
		ipHeader:   *ipHeader,
	}, nil
}

func (c *TCPConnection) Connect() error {
	addr := &syscall.SockaddrInet4{
		Port: int(c.destPort),
		Addr: c.destIP,
	}
	if err := syscall.Connect(c.rawSocket, addr); err != nil {
		return err
	}
	c.state = ESTABLISHED
	return nil
}

func (c *TCPConnection) RawConnect() error {
	synHeader := &protocol.TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum,
		ControlFlags: protocol.SYN,
		WindowSize:   65535,
		HeaderLen:    5,
	}

	log.Println("Prepare SYN packet for send")

	syncW := waiter.NewPacketChannels(c.ReceivePacket)
	syncW.StartReceive()

	time.Sleep(1 * time.Second)
	// Send SYN
	if err := c.sendPacket(synHeader); err != nil {
		return fmt.Errorf("failed to send SYN: %v", err)
	}

	log.Println("SYN packet send")

	c.state = SYN_SENT

	log.Println("Wait for SYN-ACK")

	// Wait for SYN-ACK
	resp, err := syncW.WaitForSynAck()
	if err != nil {
		return err
	}

	log.Println("prepare for send ACK")
	// Send ACK
	c.ackNum = resp.SeqNum + 1
	c.seqNum = c.seqNum + 1

	ackHeader := &protocol.TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		AckNum:       c.ackNum,
		SeqNum:       c.seqNum,
		ControlFlags: protocol.ACK,
		WindowSize:   65535,
		HeaderLen:    5,
	}

	log.Printf("ACK header")
	if err := c.sendPacket(ackHeader); err != nil {
		return fmt.Errorf("failed to send ACK: %v", err)
	}

	log.Println("send ACK")
	c.state = ESTABLISHED
	return nil
}

func (c *TCPConnection) SendMessage(data []byte) error {
	if c.state != ESTABLISHED {
		return fmt.Errorf("connection is not established")
	}

	dataHeader := &protocol.TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		AckNum:       c.ackNum,
		SeqNum:       c.seqNum,
		ControlFlags: protocol.PSH | protocol.ACK,
		WindowSize:   0xffff,
		HeaderLen:    51,
	}

	if err := c.sendPacketWithPayload(dataHeader, data); err != nil {
		return fmt.Errorf("failed to send packet with payload: %v", err)
	}

	return nil
}

func (c *TCPConnection) Close() error {
	log.Println("-----CLOSE CONN-----")
	if c.state != ESTABLISHED {
		return fmt.Errorf("connection is not established")
	}

	// Send FIN
	finHeader := &protocol.TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum + 1,
		AckNum:       c.ackNum,
		ControlFlags: protocol.FIN,
		WindowSize:   65535,
		HeaderLen:    5,
	}

	waitC := waiter.NewPacketChannels(c.ReceivePacket)
	waitC.StartReceive()
	waitF := waiter.NewPacketChannels(c.ReceivePacket)

	if err := c.sendPacket(finHeader); err != nil {
		return fmt.Errorf("failed to send FIN: %v", err)
	}

	_, err := waitC.WaitForAck()
	if err != nil {
		return fmt.Errorf("failed to receive ACK for FIN: %v", err)
	}
	waitF.StartReceive()

	resp, err := waitC.WaitForFin()
	if err != nil {
		return fmt.Errorf("failed to receive FIN: %v", err)
	}

	if err := syscall.Close(c.rawSocket); err != nil {
		return fmt.Errorf("failed to close socket: %v", err)
	}

	c.seqNum++
	c.ackNum = resp.SeqNum + 1

	// Send ACK
	ackHeader := &protocol.TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum,
		AckNum:       c.ackNum,
		ControlFlags: protocol.ACK,
		WindowSize:   65535,
		HeaderLen:    5,
	}

	if err := c.sendPacket(ackHeader); err != nil {
		return fmt.Errorf("failed to send final ACK: %v", err)
	}

	c.state = CLOSED
	return nil
}
