package message

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/tanishiking/btcwallet/protocol/common"
)

const (
	// InvTypeError means inv type ERROR.
	InvTypeError = uint32(0)

	// InvTypeMsgTx means inv type MSG_TX.
	InvTypeMsgTx = uint32(1)

	// InvTypeMsgBlock means inv type MSG_BLOCK.
	InvTypeMsgBlock = uint32(2)

	// InvTypeMsgFilteredBlock means inv type MSG_FILTERED_BLOCK.
	InvTypeMsgFilteredBlock = uint32(3)

	// InvTypeMsgCmpctBlock means inv type MSG_CMPCT_BLOCK.
	InvTypeMsgCmpctBlock = uint32(4)

	// InvvectSize means invvect's byte size.
	InvvectSize = 36
)

// Inv means inv message.
type Inv struct {
	Count     *common.VarInt
	Inventory []*InvVect
}

// NewInv create new inv.
func NewInv(count *common.VarInt, inventory []*InvVect) *Inv {
	return &Inv{
		Count:     count,
		Inventory: inventory,
	}
}

// DecodeInv deocode byte to inv.
func DecodeInv(b []byte) (*Inv, error) {
	inventory := []*InvVect{}
	varint, err := common.DecodeVarInt(b)
	if err != nil {
		return nil, err
	}
	length := len(varint.Encode())
	if uint64(len(b[length:])) != uint64(InvvectSize)*varint.Data {
		return nil, fmt.Errorf("Decode to Inv failed, invalid input: %v", b)
	}
	b = b[length:]
	for i := 0; uint64(i) < varint.Data; i++ {
		invvect, err := DecodeInvVect(b[i*InvvectSize : (i+1)*InvvectSize])
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, invvect)
	}
	return &Inv{
		Count:     varint,
		Inventory: inventory,
	}, nil
}

// CommandName return "inv".
func (inv *Inv) CommandName() string {
	return "inv"
}

// Encode encode inv.
func (inv *Inv) Encode() []byte {
	inventoryBytes := [][]byte{}
	for _, invvect := range inv.Inventory {
		inventoryBytes = append(inventoryBytes, invvect.Encode())
	}
	return bytes.Join([][]byte{
		inv.Count.Encode(),
		bytes.Join(inventoryBytes, []byte{}),
	}, []byte{})
}

// InvVect mean inv_vect.
type InvVect struct {
	InvType uint32
	Hash    [32]byte
}

// NewInvVect create new inv_vect.
func NewInvVect(invType uint32, hash [32]byte) *InvVect {
	return &InvVect{
		InvType: invType,
		Hash:    hash,
	}
}

// DecodeInvVect decode byte slice to InvVect
func DecodeInvVect(b []byte) (*InvVect, error) {
	if len(b) != InvvectSize {
		return nil, fmt.Errorf("Decode to InvVect failed, invalid input: %v", b)
	}
	var arr [32]byte
	copy(arr[:], b[4:36])
	return &InvVect{
		InvType: binary.LittleEndian.Uint32(b[0:4]),
		Hash:    arr,
	}, nil
}

// Encode inv_vect to bytes.
func (vect *InvVect) Encode() []byte {
	invType := make([]byte, 4)
	binary.LittleEndian.PutUint32(invType, vect.InvType)
	return bytes.Join([][]byte{
		invType,
		vect.Hash[:],
	}, []byte{})
}
