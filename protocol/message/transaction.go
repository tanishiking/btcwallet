package message

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/tanishiking/btcwallet/protocol/common"
	"github.com/tanishiking/btcwallet/util"
)

// TxID means transaction's ID
type TxID = [32]byte

// Transaction means bitcoin network transaction.
type Transaction struct {
	Version    uint32
	TxInCount  *common.VarInt
	TxIn       []*TxIn
	TxOutCount *common.VarInt
	TxOut      []*TxOut
	LockTime   uint32
}

// NewTransaction create new transaction.
func NewTransaction(version uint32, txIn []*TxIn, txOut []*TxOut, lockTime uint32) *Transaction {
	return &Transaction{
		Version:    version,
		TxInCount:  common.NewVarInt(uint64(len(txIn))),
		TxIn:       txIn,
		TxOutCount: common.NewVarInt(uint64(len(txOut))),
		TxOut:      txOut,
		LockTime:   lockTime,
	}
}

// DecodeTransaction decode byte slice to Transaction.
func DecodeTransaction(b []byte) (*Transaction, error) {
	version := binary.LittleEndian.Uint32(b[0:4])
	b = b[4:]

	txInArr := []*TxIn{}
	txInCount, err := common.DecodeVarInt(b)
	if err != nil {
		return nil, err
	}
	b = b[len(txInCount.Encode()):]
	for i := 0; uint64(i) < txInCount.Data; i++ {
		txIn, err := DecodeTxIn(b)
		if err != nil {
			return nil, err
		}
		txInArr = append(txInArr, txIn)
		len := len(txIn.Encode())
		b = b[len:]
	}

	txOutArr := []*TxOut{}
	txOutCount, err := common.DecodeVarInt(b)
	if err != nil {
		return nil, err
	}
	b = b[len(txOutCount.Encode()):]
	for i := 0; uint64(i) < txOutCount.Data; i++ {
		txOut, err := DecodeTxOut(b)
		if err != nil {
			return nil, err
		}
		txOutArr = append(txOutArr, txOut)
		len := len(txOut.Encode())
		b = b[len:]
	}
	if len(b) != 4 {
		return nil, fmt.Errorf("decode Transaction failed, invalid input: %v", b)
	}
	lockTime := binary.LittleEndian.Uint32(b[0:4])
	return &Transaction{
		Version:    version,
		TxInCount:  txInCount,
		TxIn:       txInArr,
		TxOutCount: txOutCount,
		TxOut:      txOutArr,
		LockTime:   lockTime,
	}, nil
}

// HasOutPoint checks the transaction has outpoint as the tx's previous output.
func (tx *Transaction) HasOutPoint(op *OutPoint) bool {
	for _, txIn := range tx.TxIn {
		if bytes.Equal(txIn.PreviousOutput.Encode(), op.Encode()) {
			return true
		}
	}
	return false
}

// FindP2khIndex find the txout which has same fromPubKeyHashed and return it's index.
func (tx *Transaction) FindP2khIndex(fromPubKeyHashed []byte) (int, error) {
	// fromPubKeyHashedと同じ宛先のpkScriptを持つtxOutのvalueのindexを返す
	p2khHeader := bytes.Join([][]byte{
		[]byte{common.OpDup},
		[]byte{common.OpHash160},
		common.NewVarStr(fromPubKeyHashed).Encode(),
	}, []byte{})
	// fmt.Println("header: ", hex.EncodeToString(p2khHeader))
	for i, txOut := range tx.TxOut {
		if bytes.HasPrefix(txOut.PkScript.Data, p2khHeader) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("No txOut matched to input: %v", fromPubKeyHashed)
}

// ID return Transaction id
func (tx *Transaction) ID() TxID {
	var res [32]byte
	hash := util.Hash256(tx.Encode())
	copy(res[:], hash)
	return res
}

// Encode encode the transaction.
func (tx *Transaction) Encode() []byte {
	versionBytes := make([]byte, 4)
	lockTimeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, tx.Version)
	binary.LittleEndian.PutUint32(lockTimeBytes, tx.LockTime)

	txInBytes := [][]byte{}
	for _, in := range tx.TxIn {
		txInBytes = append(txInBytes, in.Encode())
	}

	txOutBytes := [][]byte{}
	for _, out := range tx.TxOut {
		txOutBytes = append(txOutBytes, out.Encode())
	}

	return bytes.Join([][]byte{
		versionBytes,
		tx.TxInCount.Encode(),
		bytes.Join(txInBytes, []byte{}),
		tx.TxOutCount.Encode(),
		bytes.Join(txOutBytes, []byte{}),
		lockTimeBytes,
	}, []byte{})
}

// CommandName return message's command name.
func (tx *Transaction) CommandName() string {
	return "tx"
}

// OutPoint means transacton outpoint.
type OutPoint struct {
	Hash  TxID
	Index uint32
}

// Encode encode outpoint
func (p *OutPoint) Encode() []byte {
	indexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(indexBytes, p.Index)
	return bytes.Join([][]byte{
		p.Hash[:],
		indexBytes,
	}, []byte{})

}

// TxIn means transaction input.
type TxIn struct {
	PreviousOutput  *OutPoint
	SignatureScript *common.VarStr
	Sequence        uint32
}

// DecodeTxIn decode byte slice to transaction input.
func DecodeTxIn(b []byte) (*TxIn, error) {
	var hash [32]byte
	copy(hash[:], b[0:32])
	index := binary.LittleEndian.Uint32(b[32:36])
	out := &OutPoint{
		Hash:  hash,
		Index: index,
	}
	b = b[36:]
	signatureScript, err := common.DecodeVarStr(b)
	if err != nil {
		return nil, err
	}
	length := len(signatureScript.Encode())
	b = b[length:]
	sequence := binary.LittleEndian.Uint32(b[:4])
	return &TxIn{
		PreviousOutput:  out,
		SignatureScript: signatureScript,
		Sequence:        sequence,
	}, nil
}

// Encode encdoe TxIn to byte slice.
func (in *TxIn) Encode() []byte {
	sequenceBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sequenceBytes, in.Sequence)
	return bytes.Join([][]byte{
		in.PreviousOutput.Encode(),
		in.SignatureScript.Encode(),
		sequenceBytes,
	}, []byte{})
}

// TxOut means transaction output.
type TxOut struct {
	Value    uint64
	PkScript *common.VarStr
}

// DecodeTxOut decode byte slice to TxOut
func DecodeTxOut(b []byte) (*TxOut, error) {
	value := binary.LittleEndian.Uint64(b[0:8])
	pkScript, _ := common.DecodeVarStr(b[8:])
	return &TxOut{
		Value:    value,
		PkScript: pkScript,
	}, nil
}

// Encode encode TxOut to byte slice.
func (out *TxOut) Encode() []byte {
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, out.Value)
	return bytes.Join([][]byte{
		valueBytes,
		out.PkScript.Encode(),
	}, []byte{})
}
