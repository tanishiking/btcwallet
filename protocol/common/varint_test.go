package common

import (
	"bytes"
	"testing"
)

func TestDecodeVarIntLen1(t *testing.T) {
	b := []byte{0x6a, 0x00}
	u, _ := DecodeVarInt(b)
	length := len(u.Encode())
	expected := uint64(106)
	if u.Data != expected {
		t.Errorf("expected: %x, actual: %x", expected, u.Data)
	}
	if length != 1 {
		t.Errorf("Wrong Variable Length Integer length, expected: %d, actual: %d", 1, length)
	}
}

func TestDecodeVarIntLen3(t *testing.T) {
	b := []byte{0xfd, 0x26, 0x02, 0x00}
	u, _ := DecodeVarInt(b)
	length := len(u.Encode())
	expected := uint64(550)
	if u.Data != expected {
		t.Errorf("expected: %x, actual: %x", expected, u.Data)
	}
	if length != 3 {
		t.Errorf("Wrong Variable Length Integer length, expected: %d, actual: %d", 3, length)
	}
}

func TestDecodeVarIntLen5(t *testing.T) {
	b := []byte{0xfe, 0x70, 0x3a, 0x0f, 0x00, 0xff}
	u, _ := DecodeVarInt(b)
	length := len(u.Encode())
	expected := uint64(998000)
	if u.Data != expected {
		t.Errorf("expected: %x, actual: %x", expected, u.Data)
	}
	if length != 5 {
		t.Errorf("Wrong Variable Length Integer length, expected: %d, actual: %d", 5, length)
	}
}

func TestEncode(t *testing.T) {
	u := uint64(0xfc)
	varint := NewVarInt(u)
	expected := []byte{0xfc}
	if !bytes.Equal(varint.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varint.Encode())
	}

	zero := uint64(0)
	varintZero := NewVarInt(zero)
	expectedZero := []byte{0x00}
	if !bytes.Equal(varintZero.Encode(), expectedZero) {
		t.Errorf("expected: %x, actual: %x", expectedZero, varintZero.Encode())
	}
}

func TestPutVarIntLen3(t *testing.T) {
	u := uint64(0xfd)
	varint := NewVarInt(u)
	expected := []byte{0xfd, 0xfd, 0x00}
	if !bytes.Equal(varint.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varint.Encode())
	}
}

func TestPutVarIntLen5(t *testing.T) {
	u := uint64(0xfffffffe)
	varint := NewVarInt(u)
	expected := []byte{0xfe, 0xfe, 0xff, 0xff, 0xff}
	if !bytes.Equal(varint.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varint.Encode())
	}
}

func TestPutVarIntLen9(t *testing.T) {
	u := uint64(0xffffffffffffffff)
	varint := NewVarInt(u)
	expected := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if !bytes.Equal(varint.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varint.Encode())
	}
}
