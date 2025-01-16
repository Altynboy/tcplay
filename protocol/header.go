package protocol

import "encoding/binary"

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

func (h *TCPHeader) Serialize() []byte {
	header := make([]byte, 20) // Minimum TCP header size

	// Fill in the fields
	binary.BigEndian.PutUint16(header[0:2], h.SourcePort)
	binary.BigEndian.PutUint16(header[2:4], h.DestPort)
	binary.BigEndian.PutUint32(header[4:8], h.SeqNum)
	binary.BigEndian.PutUint32(header[8:12], h.AckNum)

	// Data offset (5 for no options) and flags
	header[12] = h.HeaderLen << 4 // Data offset (5 * 4 = 20 bytes header)
	header[13] = byte(h.ControlFlags)

	binary.BigEndian.PutUint16(header[14:16], h.WindowSize)
	binary.BigEndian.PutUint16(header[16:18], 0) // Checksum placeholder
	binary.BigEndian.PutUint16(header[18:20], h.UrgentPtr)

	// Compute checksum (pseudo-header + TCP header + payload)
	// checksum := computeChecksum(header)
	// binary.BigEndian.PutUint16(header[16:18], checksum)

	return header
}
