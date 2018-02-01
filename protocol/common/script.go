package common

import (
	"bytes"
	"encoding/binary"
	"math"
)

const (
	// OpDup mean duplicate top two data from the stack.
	OpDup = 0x76

	// OpEqual check top two data on the stack's equality.
	OpEqual = 0x87

	// OpEqualVerify check top two data on the stack's equality.
	OpEqualVerify = 0x88

	// OpHash160 hash256 and then ripemd on the top of stack.
	OpHash160 = 0xa9

	// OpCheckSig checks signature.
	OpCheckSig = 0xac
)

// OpPushData return script to push data.
// https://en.bitcoin.it/wiki/Script#Opcodes
func OpPushData(data []byte) []byte {
	len := len(data)
	if len <= 75 {
		return bytes.Join([][]byte{
			[]byte{byte(len)},
			data,
		}, []byte{})
	}
	if len <= math.MaxUint8 {
		return bytes.Join([][]byte{
			[]byte{0x4c},
			[]byte{byte(len)},
			data,
		}, []byte{})
	}
	if len <= math.MaxUint16 {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(len))
		return bytes.Join([][]byte{
			[]byte{0x4d},
			b,
			data,
		}, []byte{})
	}
	if len <= math.MaxUint32 {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(len))
		return bytes.Join([][]byte{
			[]byte{0x4e},
			b,
			data,
		}, []byte{})
	}
	// TODO: error if data is too large
	return []byte{}
}
