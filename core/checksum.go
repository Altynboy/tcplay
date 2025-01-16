package core

import (
	"encoding/binary"
	"fmt"
	"tcplay/protocol"
)

// 0      7 8     15 16    23 24    31
// +--------+--------+--------+--------+
// |          Source Address           |
// +--------+--------+--------+--------+
// |        Destination Address        |
// +--------+--------+--------+--------+
// |  Zero  |Protocol|   TCP Length    |
// +--------+--------+--------+--------+

type IPPseudoHeader struct {
	SourceIP [4]byte
	DestIP   [4]byte
	Zero     uint8  // padding required by the TCP/IP spec
	Protocol uint8  // 6 for TCP
	Length   uint16 // TCP header + data length
}

func (c *TCPConnection) calculateChecksum(header *protocol.TCPHeader, data []byte, srcIP [4]byte, destIP [4]byte) uint16 {
	if header.Checksum != 0 {
		fmt.Println("TCP header should be with zero checksum")
		return 0
	}

	// 1. Create pseudo IP header
	pseudoHeader := &IPPseudoHeader{
		SourceIP: srcIP,
		DestIP:   destIP,
		Protocol: 6,
		Length:   uint16(20 + len(data)),
	}

	// 2. Calculate total size and create buffer
	totalLen := 12 + 20 + len(data) // pseudo-header + TCP header + data
	if len(data)&2 != 0 {
		totalLen++
	}
	buf := make([]byte, totalLen)

	// 3. Copy pseudo header
	binary.BigEndian.PutUint32(buf[0:4], binary.BigEndian.Uint32(pseudoHeader.SourceIP[:]))
	binary.BigEndian.PutUint32(buf[4:8], binary.BigEndian.Uint32(pseudoHeader.DestIP[:]))
	buf[8] = pseudoHeader.Zero
	buf[9] = pseudoHeader.Protocol
	binary.BigEndian.PutUint16(buf[10:12], pseudoHeader.Length)

	// 4. Copy TCP header (with zero checksum)
	offset := 12
	binary.BigEndian.AppendUint16(buf[offset:offset+2], header.SourcePort)
	binary.BigEndian.AppendUint16(buf[offset+2:offset+4], header.DestPort)
	binary.BigEndian.AppendUint32(buf[offset+4:offset+8], header.SeqNum)
	binary.BigEndian.AppendUint32(buf[offset+8:offset+12], header.AckNum)
	buf[offset+12] = header.HeaderLen << 4
	buf[offset+13] = header.ControlFlags
	binary.BigEndian.AppendUint16(buf[offset+14:offset+16], header.WindowSize)
	binary.BigEndian.AppendUint16(buf[offset+16:offset+18], 0)
	binary.BigEndian.AppendUint16(buf[offset+18:offset+20], header.UrgentPtr)

	if len(data) > 0 {
		copy(buf[offset+20:], data)
	}

	var sum uint32
	for i := 0; i < len(buf)-1; i += 2 {
		sum += uint32(buf[i])<<8 | uint32(buf[i+1])
	}

	// 0xffff - 65535
	if sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	return ^uint16(sum)
}
