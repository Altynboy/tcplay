package core

func (c *TCPConnection) sendPacket(header *TCPHeader) error {

	return nil
}

func (c *TCPConnection) receivePacket() (*TCPHeader, error) {
	return nil, nil
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
