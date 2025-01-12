package core

import (
	"fmt"
	"log"
	"syscall"

	"math/rand"
)

type TCPHeader struct {
	SourcePort   uint16
	DestPort     uint16
	SeqNum       uint32
	AckNum       uint32
	HeaderLen    uint8 // 4 bit field
	ControlFlags uint8
	WindowSize   uint16
	Checksum     uint16
	UrgentPtr    uint16
}

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
}

const (
	ACK = 16 // 0001 0000
	PSH = 8  // 0000 1000
	RST = 4  // 0000 0100
	SYN = 2  // 0000 0010
	FIN = 1  // 0000 0001
)

const (
	CLOSED      = 0
	SYN_SENT    = 1
	ESTABLISHED = 2
)

func CreateConnection(destPort uint16, destIP [4]byte) (*TCPConnection, error) {
	sourcePort := uint16(49152 + rand.Intn(65535-49152+1))
	srcIP := [4]byte{127, 0, 0, 1}

	// Create raw socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %v", err)
	}

	return &TCPConnection{
		srcPort:    sourcePort,
		destPort:   destPort,
		srcIP:      srcIP,
		destIP:     destIP,
		seqNum:     generateRandomSeqNum(),
		ackNum:     0,
		state:      CLOSED,
		rawSocket:  fd,
		maxSegSize: 1460,
	}, nil
}

func (c *TCPConnection) Connect() error {
	synHeader := &TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum,
		ControlFlags: SYN,
		WindowSize:   65535,
		HeaderLen:    5,
	}

	log.Println("Prepare SYN packet for send")

	// Send SYN
	if err := c.sendPacket(synHeader); err != nil {
		return fmt.Errorf("failed to send SYN: %v", err)
	}

	log.Println("SYN packet send")

	c.state = SYN_SENT

	log.Println("Wait for SYN-ACK")
	// Wait for SYN-ACK
	resp, err := c.receivePacket()
	if err != nil {
		return fmt.Errorf("failed to receive SYN-ACK: %v", err)
	}
	if resp.ControlFlags != (SYN | ACK) {
		return fmt.Errorf("expected SYN-ACK, got different flags: %d", resp.ControlFlags)
	}

	log.Println("prepare for send ACK")
	// Send ACK
	c.ackNum = resp.AckNum + 1
	c.seqNum = c.seqNum + 1

	ackHeader := &TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		AckNum:       c.ackNum,
		SeqNum:       c.seqNum,
		ControlFlags: ACK,
		WindowSize:   65535,
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

	dataHeader := &TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		AckNum:       c.ackNum,
		SeqNum:       c.seqNum,
		ControlFlags: PSH | ACK,
		WindowSize:   0xffff,
	}

	if err := c.sendPacketWithPayload(dataHeader, data); err != nil {
		return fmt.Errorf("failed to send pavket with payload: %v", err)
	}

	return nil
}

func (c *TCPConnection) Close() error {
	if c.state != ESTABLISHED {
		return fmt.Errorf("connection is not established")
	}

	// Send FIN
	finHeader := &TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum,
		AckNum:       c.ackNum,
		ControlFlags: FIN,
		WindowSize:   65535,
	}

	if err := c.sendPacket(finHeader); err != nil {
		return fmt.Errorf("failed to send FIN: %v", err)
	}

	// Wait for ACK and FIN
	if err := c.waitForACK(); err != nil {
		return fmt.Errorf("failed to receive ACK for FIN: %v", err)
	}

	if err := c.waitForFIN(); err != nil {
		return fmt.Errorf("failed to receive FIN: %v", err)
	}

	if err := syscall.Close(c.rawSocket); err != nil {
		return fmt.Errorf("failed to close socket: %v", err)
	}

	// Send ACK
	ackHeader := &TCPHeader{
		SourcePort:   c.srcPort,
		DestPort:     c.destPort,
		SeqNum:       c.seqNum,
		AckNum:       c.ackNum + 1,
		ControlFlags: ACK,
		WindowSize:   65535,
	}

	if err := c.sendPacket(ackHeader); err != nil {
		return fmt.Errorf("failed to send final ACK: %v", err)
	}

	c.state = CLOSED
	return nil
}
