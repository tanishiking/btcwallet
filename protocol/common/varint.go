package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// VarInt means variable length integer
// refer: https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
type VarInt struct {
	Data uint64
}

// NewVarInt return new VarInt.
func NewVarInt(u uint64) *VarInt {
	return &VarInt{
		Data: u,
	}
}

// DecodeVarInt decode byte slice to variable length integer.
// https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
func DecodeVarInt(bs []byte) (*VarInt, error) {
	if bytes.HasPrefix(bs, []byte{0xff}) {
		return &VarInt{
			Data: binary.LittleEndian.Uint64(bs[1:9]),
		}, nil
	}
	if bytes.HasPrefix(bs, []byte{0xfe}) {
		return &VarInt{
			Data: uint64(binary.LittleEndian.Uint32(bs[1:5])),
		}, nil
	}
	if bytes.HasPrefix(bs, []byte{0xfd}) {
		return &VarInt{
			Data: uint64(binary.LittleEndian.Uint16(bs[1:3])),
		}, nil
	}
	if bytes.Compare(bs, []byte{0xfd}) < 0 {
		if len(bs) == 0 {
			return nil, fmt.Errorf("Decode VarInt failed, invalid input: %v", bs)
		}
		return &VarInt{Data: uint64(bs[0])}, nil
	}
	return nil, fmt.Errorf("Decode VarInt failed, invalid input: %v", bs)
}

// Encode convert uint64 to bytes formatted in byc variable length integer
// https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
func (v *VarInt) Encode() []byte {
	if v.Data < 0xfd {
		return []byte{byte(v.Data)}
	}
	if v.Data <= 0xffff {
		b := make([]byte, 3)
		b[0] = byte(0xfd)
		binary.LittleEndian.PutUint16(b[1:], uint16(v.Data))
		return b
	}
	if v.Data <= 0xffffffff {
		b := make([]byte, 5)
		b[0] = byte(0xfe)
		binary.LittleEndian.PutUint32(b[1:], uint32(v.Data))
		return b
	}
	if v.Data <= 0xffffffffffffffff {
		b := make([]byte, 9)
		b[0] = byte(0xff)
		binary.LittleEndian.PutUint64(b[1:], v.Data)
		return b
	}
	return []byte{byte(v.Data)}
}
