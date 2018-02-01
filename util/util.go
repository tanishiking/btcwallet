package util

import (
	"bytes"
	"crypto/sha256"
	"io"
	"math/rand"
	"time"

	"golang.org/x/crypto/ripemd160"
)

// Hash256 perform SHA-256 hash on the data twice.
func Hash256(data []byte) []byte {
	s := sha256.Sum256(data)
	s2 := sha256.Sum256(s[:])
	return s2[:]
	// return sha256.Sum256(sha256.Sum256(data))
}

// Hash160 perform SHA-256 hash on the data and then perform RIPEMD-160 hash.
//
// refer: https://en.bitcoin.it/wiki/RIPEMD-160
func Hash160(data []byte) []byte {
	rip := ripemd160.New()
	sum := sha256.Sum256(data)
	io.WriteString(rip, string(sum[:]))
	return rip.Sum(nil)
}

// RandInt generate random utin8 integer.
func RandInt(min int, max int) uint8 {
	rand.Seed(time.Now().UTC().UnixNano())
	return uint8(min + rand.Intn(max-min))
}

// ReverseBytes reverse byte slice.
func ReverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

// SubTxIDs return IDs1 sub IDs2
func SubTxIDs(IDs1 [][32]byte, IDs2 [][32]byte) [][32]byte {
	diff := [][32]byte{}
	for _, ID1 := range IDs1 {
		found := false
		for _, ID2 := range IDs2 {
			if bytes.Equal(ID1[:], ID2[:]) {
				found = true
			}
		}
		if !found {
			diff = append(diff, ID1)
		}
	}
	return diff
}
