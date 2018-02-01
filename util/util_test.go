package util

import (
	"bytes"
	"math"
	"testing"
)

func TestReverseBytes(t *testing.T) {
	bs1 := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	expected1 := []byte{0x04, 0x03, 0x02, 0x01, 0x00}
	reversed1 := ReverseBytes(bs1)
	if !bytes.Equal(reversed1, expected1) {
		t.Errorf("expected: %v, actual: %v", expected1, reversed1)
	}

	bs2 := []byte{0x00, 0x01, 0x02, 0x03}
	expected2 := []byte{0x03, 0x02, 0x01, 0x00}
	reversed2 := ReverseBytes(bs2)
	if !bytes.Equal(reversed2, expected2) {
		t.Errorf("expected: %v, actual: %v", expected2, reversed2)
	}
}

func TestSubTxIDs(t *testing.T) {
	id1 := genDummyTxID()
	id2 := genDummyTxID()
	id3 := genDummyTxID()
	res := SubTxIDs([][32]byte{id1, id2, id3}, [][32]byte{id1, id3})
	if len(res) != 1 {
		t.Errorf("Test SubTxIDs failed, res should be %v, actual: %v", 1, len(res))
	}
	txID := res[0]
	if !bytes.Equal(txID[:], id2[:]) {
		t.Errorf("expected: %v, actual: %v", id2[:], txID[:])
	}
}

func genDummyTxID() [32]byte {
	var res [32]byte
	for i := 0; i < 32; i++ {
		res[i] = RandInt(0, math.MaxUint8)
	}
	return res
}
