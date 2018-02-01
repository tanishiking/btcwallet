package common

import (
	"bytes"
	"fmt"
)

// VarStr means variable length string
// refer: https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_string
type VarStr struct {
	VarInt *VarInt
	Data   []byte
}

// NewVarStr create new VarStr from data.
func NewVarStr(b []byte) *VarStr {
	len := uint64(len(b))
	varint := NewVarInt(len)
	return &VarStr{
		VarInt: varint,
		Data:   b,
	}
}

// DecodeVarStr decode byte to VarStr.
func DecodeVarStr(b []byte) (*VarStr, error) {
	varint, err := DecodeVarInt(b)
	if err != nil {
		return nil, err
	}
	varintLen := len(varint.Encode())
	varstrLen := varint.Data + uint64(varintLen)
	if uint64(len(b)) < varstrLen {
		return nil, fmt.Errorf("Decode varstr failed, invalid input: %v", b)
	}
	str := b[varintLen:varstrLen]
	return &VarStr{
		VarInt: varint,
		Data:   str,
	}, nil
}

// Encode encode VarStr to byte slice.
func (s *VarStr) Encode() []byte {
	return bytes.Join([][]byte{
		s.VarInt.Encode(),
		s.Data,
	},
		[]byte{},
	)
}
