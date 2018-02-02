package protocol

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"

	secp256k1 "github.com/toxeus/go-secp256k1"

	"github.com/tanishiking/btcwallet/key"
	"github.com/tanishiking/btcwallet/protocol/common"
	"github.com/tanishiking/btcwallet/protocol/message"
	"github.com/tanishiking/btcwallet/util"
)

// Send send bitcoint to toAddr with amount and fee.
func Send(toAddr string, amount int, fee int) {
	fn := func(conn net.Conn, v *message.Version) {
		utxos := collectUTXO(conn, v)
		value := uint64(0)
		utxoInput := []*utxo{}
		for _, unspent := range utxos {
			utxoInput = append(utxoInput, unspent)
			value += unspent.tx.TxOut[unspent.index].Value
			if uint64(amount+fee) <= value {
				break
			}
		}
		if value < uint64(amount+fee) {
			fmt.Printf("Balance is not enough, balance: %v, amount: %v, fee: %v\n", value, amount, fee)
			os.Exit(1)
		}
		txOut, err := createTxOut(toAddr, amount, value, fee)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		txIn, err := createTxIn(utxoInput, txOut)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		transaction := message.NewTransaction(uint32(1), txIn, txOut, uint32(0))

		inv := message.NewInv(
			common.NewVarInt(uint64(1)),
			[]*message.InvVect{message.NewInvVect(message.InvTypeMsgTx, transaction.ID())},
		)
		SendMessage(conn, inv)

		var header [common.MessageHeaderLen]byte
		buf := make([]byte, common.MessageHeaderLen)
	Loop:
		for {
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println(err.Error())
				break Loop
			}
			if n == common.MessageHeaderLen {
				copy(header[:], buf)
				mh := common.DecodeMessageHeader(header)
				msgBytes, err := RecvMessage(conn, mh.Length)
				if err != nil {
					fmt.Println(err.Error())
					break Loop
				}
				fmt.Printf("Recv: %s %d\n", string(mh.Command[:]), mh.Length)
				if bytes.HasPrefix(mh.Command[:], []byte("getdata")) {
					getData, err := message.DecodeGetData(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					invs := getData.FilterInventoryWithType(message.InvTypeMsgTx)
					for _, invvect := range invs {
						txID := transaction.ID()
						if bytes.Equal(invvect.Hash[:], txID[:]) {
							fmt.Println("transaction send!")
							SendMessage(conn, transaction)
						}
					}
				} else if bytes.HasPrefix(mh.Command[:], []byte("reject")) {
					reject, err := message.DecodeReject(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					fmt.Println(reject.String())
				}
			}
		}
	}
	WithBitcoinConnection(fn)
}

func createTxIn(unspentTxs []*utxo, txOut []*message.TxOut) ([]*message.TxIn, error) {
	fromPrivateKey, err := key.ReadOrGeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	fromPublicKey, err := key.GeneratePubKey(fromPrivateKey)
	if err != nil {
		return nil, err
	}

	res := []*message.TxIn{}

	for _, unspent := range unspentTxs {
		previoutTx := unspent.tx
		previousTxID := previoutTx.ID()
		previousOutput := previoutTx.TxOut[unspent.index]

		output := &message.OutPoint{
			Hash:  previousTxID,
			Index: unspent.index,
		}

		txCopyInput := []*message.TxIn{}
		for _, otherUnspent := range unspentTxs {
			// 署名の作成
			// https://en.bitcoin.it/wiki/OP_CHECKSIG
			// previous transaction の locking script を subscript とする
			// もしくは OP_SP がある場合は OP_SP で split して last を subscript
			if unspent.equal(otherUnspent) {
				tmpTxIn := &message.TxIn{
					PreviousOutput:  output,
					SignatureScript: previousOutput.PkScript,
					Sequence:        0xFFFFFFFF, // ignored
				}
				txCopyInput = append(txCopyInput, tmpTxIn)
			} else {
				otherTx := otherUnspent.tx
				otherTxID := otherTx.ID()

				otherOutput := &message.OutPoint{
					Hash:  otherTxID,
					Index: otherUnspent.index,
				}

				emptyScript := common.NewVarStr([]byte{})
				tmpTxIn := &message.TxIn{
					PreviousOutput:  otherOutput,
					SignatureScript: emptyScript,
					Sequence:        0xFFFFFFFF, // ignored
				}
				txCopyInput = append(txCopyInput, tmpTxIn)
			}
		}

		// これから作るトランザクションのうちinputがsubscriptなものを作る
		txCopy := message.NewTransaction(
			uint32(1),
			txCopyInput,
			txOut,
			uint32(0),
		)
		// 末尾にhashTypeCodeをつけてhash256
		hashType := []byte{0x01}
		verified := util.Hash256(bytes.Join([][]byte{
			txCopy.Encode(),
			[]byte{0x01, 0x00, 0x00, 0x00},
		}, []byte{}))

		secp256k1.Start()
		var rawTxHashed [32]byte
		var privKeyByte [32]byte
		copy(rawTxHashed[:], verified)
		copy(privKeyByte[:], fromPrivateKey)
		signedTx, ok := secp256k1.Sign(rawTxHashed, privKeyByte, nil)
		secp256k1.Stop()
		if !ok {
			return nil, fmt.Errorf("Failed to sign transaction %v", rawTxHashed)
		}
		signedTxWithType := bytes.Join([][]byte{signedTx, hashType}, []byte{})

		unlockingScript := common.NewVarStr(bytes.Join([][]byte{
			common.OpPushData(signedTxWithType),
			common.OpPushData(fromPublicKey),
		}, []byte{}))

		input := &message.TxIn{
			PreviousOutput:  output,
			SignatureScript: unlockingScript,
			Sequence:        0xFFFFFFFF, // ignored
		}
		res = append(res, input)
	}
	return res, nil
}

func createTxOut(toAddr string, amount int, balance uint64, fee int) ([]*message.TxOut, error) {
	toPubKeyHashed, err := key.DecodeBitcoinAddr(toAddr)
	if err != nil {
		return nil, err
	}
	fromPrivateKey, err := key.ReadOrGeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	fromPubKey, err := key.GeneratePubKey(fromPrivateKey)
	if err != nil {
		return nil, err
	}
	fromPubKeyHashed := util.Hash160(fromPubKey)

	// P2SH
	lockingScript1 := common.NewVarStr(bytes.Join([][]byte{
		[]byte{common.OpHash160},
		common.OpPushData(toPubKeyHashed),
		[]byte{common.OpEqual},
	}, []byte{}))

	// P2PKH
	lockingScript2 := common.NewVarStr(bytes.Join([][]byte{
		[]byte{common.OpDup},
		[]byte{common.OpHash160},
		common.OpPushData(fromPubKeyHashed),
		[]byte{common.OpEqualVerify},
		[]byte{common.OpCheckSig},
	}, []byte{}))

	txOut1 := &message.TxOut{
		Value:    uint64(amount),
		PkScript: lockingScript1,
	}
	txOut2 := &message.TxOut{
		Value:    balance - uint64(amount+fee),
		PkScript: lockingScript2,
	}
	return []*message.TxOut{txOut1, txOut2}, nil
}
