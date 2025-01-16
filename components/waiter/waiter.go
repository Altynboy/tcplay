package waiter

import (
	"fmt"
	"tcplay/protocol"
)

// Define the struct with channels
type PacketChannels struct {
	ch             chan *protocol.TCPHeader
	errCh          chan error
	receivePacketF func() (*protocol.TCPHeader, error)
}

// Method to create and initialize the channels in the struct
func NewPacketChannels(receivePacketF func() (*protocol.TCPHeader, error)) *PacketChannels {
	return &PacketChannels{
		ch:             make(chan *protocol.TCPHeader),
		errCh:          make(chan error),
		receivePacketF: receivePacketF,
	}
}

// Method to start receiving a packet in the background
func (c *PacketChannels) StartReceive() {

	go func(c *PacketChannels) {
		resp, err := c.receivePacketF()
		if err != nil {
			c.errCh <- fmt.Errorf("failed to receive packet: %v", err)
			return
		}
		c.ch <- resp
	}(c)
}

// Method to wait for the response from the channels
func (c *PacketChannels) waitForResponse() (*protocol.TCPHeader, error) {
	var resp *protocol.TCPHeader
	select {
	case resp = <-c.ch:
		return resp, nil
	case err := <-c.errCh:
		return nil, err
	}
}

// The main function to wait for the ACK packet, using the channels in the struct
func (c *PacketChannels) WaitForAck() (*protocol.TCPHeader, error) {
	// Wait for the response
	resp, err := c.waitForResponse()
	if err != nil {
		return nil, err
	}

	// Check if the response is an ACK
	if resp.ControlFlags != protocol.ACK {
		return nil, fmt.Errorf("expected ACK, got different flags: %d", resp.ControlFlags)
	}

	return resp, nil
}

func (c *PacketChannels) WaitForSynAck() (*protocol.TCPHeader, error) {
	// Wait for the response
	resp, err := c.waitForResponse()
	if err != nil {
		return nil, err
	}

	// Check if the response is an ACK
	if resp.ControlFlags != (protocol.SYN | protocol.ACK) {
		return nil, fmt.Errorf("expected SYN-ACK, got different flags: %d", resp.ControlFlags)
	}

	return resp, nil
}

func (c *PacketChannels) WaitForFin() (*protocol.TCPHeader, error) {
	// Wait for the response
	resp, err := c.waitForResponse()
	if err != nil {
		return nil, err
	}

	// Check if the response is an ACK
	if resp.ControlFlags != protocol.FIN {
		return nil, fmt.Errorf("expected FIN, got different flags: %d", resp.ControlFlags)
	}

	return resp, nil
}
