package common

import (
	"bytes"
	"testing"
)

func TestNetAddrEncode(t *testing.T) {
	addr := &NetAddr{
		Services: uint64(1),
		IP: [16]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x7F, 0x00, 0x00, 0x01,
		},
		Port: uint16(256),
	}
	actual := addr.Encode()
	expected := bytes.Join([][]byte{
		[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		addr.IP[:],

		[]byte{0x01, 0x00},
	},
		[]byte{},
	)
	if !bytes.Equal(actual[:], expected) {
		t.Errorf("expected: %x, actual: %x", expected, actual[:])
	}
}
