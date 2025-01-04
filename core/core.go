package core

type TCPHeader struct {
	SourcePort   uint16
	DestPort     uint16
	SeqNum       uint32
	AckNum       uint32
	HeaderLen    uint8
	ControlFlags uint8
	WindowSize   uint16
	Checksum     uint16
	UrgentPtr    uint16
}

type TCPConnection struct {
	srcPort   uint16
	destPort  uint16
	destIP    [4]byte
	seqNum    uint32
	ackNum    uint32
	state     uint8
	rawSocket int
}

const (
	ACK = 16 // 0001 0000
	PSH = 8  // 0000 1000
	RST = 4  // 0000 0100
	SYN = 2  // 0000 0010
	FIN = 1  // 0000 0001
)
