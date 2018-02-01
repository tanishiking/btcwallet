package common

import (
	"bytes"
	"encoding/binary"
)

// MessageHeader means messageheader.
type MessageHeader struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum [4]byte
}

// DecodeMessageHeader decode byte array to messageheader
func DecodeMessageHeader(b [24]byte) *MessageHeader {
	var (
		command  [12]byte
		checksum [4]byte
	)
	copy(command[:], b[4:16])
	copy(checksum[:], b[20:24])
	return &MessageHeader{
		Magic:    binary.LittleEndian.Uint32(b[0:4]),
		Command:  command,
		Length:   binary.LittleEndian.Uint32(b[16:20]),
		Checksum: checksum,
	}
}

// Encode encode messageheader.
func (header *MessageHeader) Encode() []byte {
	var (
		magic  [4]byte
		length [4]byte
	)
	binary.LittleEndian.PutUint32(magic[:], header.Magic)
	binary.LittleEndian.PutUint32(length[:], header.Length)
	return bytes.Join([][]byte{
		magic[:],
		header.Command[:],
		length[:],
		header.Checksum[:],
	},
		[]byte{},
	)
}
