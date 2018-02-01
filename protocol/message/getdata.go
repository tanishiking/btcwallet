package message

import (
	"bytes"
	"fmt"

	"github.com/tanishiking/btcwallet/protocol/common"
)

// GetData is message to get inventory data.
// https://en.bitcoin.it/wiki/Protocol_documentation#getdata
type GetData struct {
	Count     *common.VarInt
	Inventory []*InvVect
}

// NewGetData create new getdata message.
func NewGetData(inventory []*InvVect) *GetData {
	length := len(inventory)
	count := common.NewVarInt(uint64(length))
	return &GetData{
		Count:     count,
		Inventory: inventory,
	}
}

// DecodeGetData decode byte to GetData struct.
func DecodeGetData(b []byte) (*GetData, error) {
	varint, err := common.DecodeVarInt(b)
	if err != nil {
		return nil, err
	}
	length := len(varint.Encode())
	if uint64(len(b)) != uint64(length)+varint.Data*uint64(InvvectSize) {
		return nil, fmt.Errorf("Decode to GetData failed, invalid input %v", b)
	}
	inventory := []*InvVect{}
	for i := 0; uint64(i) < varint.Data; i++ {
		vectByte := b[length+i*InvvectSize : length+(i+1)*InvvectSize]
		invVect, err := DecodeInvVect(vectByte)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, invVect)
	}
	return &GetData{
		Count:     varint,
		Inventory: inventory,
	}, nil
}

// CommandName return message's command name.
func (g *GetData) CommandName() string {
	return "getdata"
}

// Encode encdoe message to byte slice.
func (g *GetData) Encode() []byte {
	inventoryBytes := [][]byte{}
	for _, invvect := range g.Inventory {
		inventoryBytes = append(inventoryBytes, invvect.Encode())
	}
	return bytes.Join([][]byte{
		g.Count.Encode(),
		bytes.Join(inventoryBytes, []byte{}),
	}, []byte{})
}

// FilterInventoryWithType return filtered InvVects by inv type.
func (g *GetData) FilterInventoryWithType(typ uint32) []*InvVect {
	inventory := []*InvVect{}
	for _, invvect := range g.Inventory {
		if invvect.InvType == typ {
			inventory = append(inventory, invvect)
		}
	}
	return inventory
}
