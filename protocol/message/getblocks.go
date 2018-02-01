package message

import (
	"bytes"
	"encoding/binary"

	"github.com/tanishiking/btcwallet/protocol/common"
)

// ZeroHash means hashstop that request to get as many blocks as possible (500)
var ZeroHash = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// Getblocks means getblocks message.
type Getblocks struct {
	Version            uint32
	HashCount          *common.VarInt
	BlockLocatorHashes [][32]byte
	HashStop           [32]byte
}

// NewGetBlocks create new Getblocks.
func NewGetBlocks(version uint32, blockLocatorHashes [][32]byte, hashStop [32]byte) *Getblocks {
	length := len(blockLocatorHashes)
	hashCount := common.NewVarInt(uint64(length))
	return &Getblocks{
		Version:            version,
		HashCount:          hashCount,
		BlockLocatorHashes: blockLocatorHashes,
		HashStop:           hashStop,
	}
}

// CommandName return "getblocks"
func (g *Getblocks) CommandName() string {
	return "getblocks"
}

// Encode encode getblocks.
func (g *Getblocks) Encode() []byte {
	versionByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionByte, g.Version)
	hashesBytes := [][]byte{}
	for _, hash := range g.BlockLocatorHashes {
		hashesBytes = append(hashesBytes, hash[:])
	}
	return bytes.Join([][]byte{
		versionByte,
		g.HashCount.Encode(),
		bytes.Join(hashesBytes, []byte{}),
		g.HashStop[:],
	}, []byte{})
}
