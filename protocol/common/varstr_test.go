package common

import (
	"bytes"
	"testing"
)

func TestVarStr(t *testing.T) {
	str := []byte("gopher")
	// []byte{0x67, 0x6f, 0x70, 0x68, 0x65, 0x72}
	expected := bytes.Join([][]byte{[]byte{0x06}, []byte(str)}, []byte{})
	varstr := NewVarStr(str)
	if !bytes.Equal(varstr.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varstr.Encode())
	}
}

func TestVarStrEmpty(t *testing.T) {
	str := []byte{}
	expected := []byte{0x00}
	varstr := NewVarStr(str)
	if !bytes.Equal(varstr.Encode(), expected) {
		t.Errorf("expected: %x, actual: %x", expected, varstr.Encode())
	}
}
