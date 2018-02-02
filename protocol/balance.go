package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/tanishiking/btcwallet/key"
	"github.com/tanishiking/btcwallet/protocol/common"
	"github.com/tanishiking/btcwallet/protocol/message"
	"github.com/tanishiking/btcwallet/util"
)

type utxo struct {
	tx    *message.Transaction
	index uint32
}

func (u *utxo) equal(other *utxo) bool {
	return bytes.Equal(u.tx.Encode(), other.tx.Encode()) && u.index == other.index
}

// Balance show the balance of this wallet.
func Balance() {
	fn := func(conn net.Conn, v *message.Version) {
		utxos := collectUTXO(conn, v)
		balance := uint64(0)
		for _, utxo := range utxos {
			balance += utxo.tx.TxOut[utxo.index].Value
		}
		fmt.Println("残高: ", balance)
	}
	WithBitcoinConnection(fn)
}

func collectUTXO(conn net.Conn, v *message.Version) []*utxo {
	blockCh := make(chan *message.Merkleblock)
	txCh := make(chan *message.Transaction)

	// 各種メッセージを受け取るgoroutineを立ち上げておく
	go dispatch(conn, blockCh, txCh)

	// 鍵の準備
	fromPrivateKey, err := key.ReadOrGeneratePrivateKey()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fromPublicKey, err := key.GeneratePubKey(fromPrivateKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fromPublicKeyHash := util.Hash160(fromPublicKey)

	startBlockHash, err := hex.DecodeString("0000000000000657bda6681e1a3d1aac92d09d31721e8eedbca98cac73e93226")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var arr [32]byte
	copy(arr[:], util.ReverseBytes(startBlockHash))

	leftBlocks := v.StartHeight - uint32(1261780)

	// merkleblockの送信要請のためgetblocksを送信
	SendMessage(conn, message.NewFilterload(1024, 10, [][]byte{fromPublicKeyHash}))
	getBlocksMessage := message.NewGetBlocks(uint32(70015), [][32]byte{arr}, message.ZeroHash)
	SendMessage(conn, getBlocksMessage)

	fmt.Println("left blocks: ", leftBlocks)

	txs := []*message.Transaction{}
	txRecvDoneCh := make(chan struct{})
	go getTxs(conn, txCh, txRecvDoneCh, &txs)

	// merkleblockを受信
	merkleBlocks := []*message.Merkleblock{}
	blockRecvDoneCh := make(chan struct{})
	// goroutineでmerkleblockを受信、受信完了までブロック
	go getBlocks(conn, blockCh, leftBlocks, blockRecvDoneCh, &merkleBlocks)
	<-blockRecvDoneCh

	// merkleblockからトランザクションIDを取り出す
	targetTxIDs := [][32]byte{}
	for _, m := range merkleBlocks {
		matchedTxs := m.Validate()
		for _, h := range matchedTxs {
			targetTxIDs = append(targetTxIDs, h)
		}
	}
	fmt.Println("want transactions: ", len(targetTxIDs))

	receivedTxIDs := [][32]byte{}
	for _, tx := range txs {
		receivedTxIDs = append(receivedTxIDs, tx.ID())
	}
	unkowns := util.SubTxIDs(targetTxIDs, receivedTxIDs)

	// 受け取ったTxIDからtrasactionを受信するためにgetDataを送信
	inventory := []*message.InvVect{}
	for _, h := range unkowns {
		invvect := message.NewInvVect(message.InvTypeMsgTx, h)
		inventory = append(inventory, invvect)
	}
	getData := message.NewGetData(inventory)
	SendMessage(conn, getData)

	// 受け取りたいトランザクションを全て受け取るまでループ
Loop:
	for {
		receivedTxIDs = [][32]byte{}
		for _, tx := range txs {
			receivedTxIDs = append(receivedTxIDs, tx.ID())
		}
		unkowns := util.SubTxIDs(targetTxIDs, receivedTxIDs)
		if len(unkowns) == 0 {
			txRecvDoneCh <- struct{}{}
			break Loop
		}
	}

	utxos := []*utxo{}
	for _, tx := range txs {
		txID := tx.ID()
		fmt.Println(hex.EncodeToString(txID[:]))
		index, err := tx.FindP2khIndex(fromPublicKeyHash)
		if err != nil {
			continue
		}
		fmt.Println(tx.TxOut[index].Value)
		outPoint := &message.OutPoint{
			Hash:  txID,
			Index: uint32(index),
		}
		spend := false
		for _, otherTx := range txs {
			if otherTx.HasOutPoint(outPoint) {
				spend = true
				break
			}
		}
		if !spend {
			unspent := &utxo{
				tx:    tx,
				index: uint32(index),
			}
			utxos = append(utxos, unspent)
		}
	}
	return utxos
}

func getBlocks(conn net.Conn, blockCh chan *message.Merkleblock, leftBlocks uint32, doneCh chan struct{}, blocks *[]*message.Merkleblock) {
	merkleBlocks := message.NewMerkleBlocks()

	// fmt.Println("left blocks: ", leftBlocks)

	// zeroHash を使ってgetblocksを送信した場合最大500個のmerkleblockが送信される
	bunch := 500
Loop:
	for {
		if uint32(merkleBlocks.Size()) >= leftBlocks {
			doneCh <- struct{}{}
			break Loop
		}
		if merkleBlocks.Size() >= bunch {
			// 500blocks受信したら
			// 受信したmerkleblockのうち最新のblockHashを使ってgetblocksを再度送る
			bunch += 500
			latestBlockHash := merkleBlocks.LatestBlock().BlockHash()

			getBlocksMessage := message.NewGetBlocks(uint32(70015), [][32]byte{latestBlockHash}, message.ZeroHash)
			SendMessage(conn, getBlocksMessage)
		}
		select {
		case mb := <-blockCh:
			fmt.Println(merkleBlocks.Size())
			merkleBlocks.Add(mb)
			*blocks = append(*blocks, mb)
		case <-time.After(time.Second * 10):
			// 途中で詰まることがよくあるので
			// 3秒merkleblockが送られて来なかった場合は再度getblocksを送信
			bunch = merkleBlocks.Size() + 500
			latestBlock := merkleBlocks.LatestBlock()
			if latestBlock != nil {
				latestBlockHash := latestBlock.BlockHash()
				getBlocksMessage := message.NewGetBlocks(uint32(70015), [][32]byte{latestBlockHash}, message.ZeroHash)
				SendMessage(conn, getBlocksMessage)
			} else {
				doneCh <- struct{}{}
				break Loop
			}
		}
	}
	for {
		// 以降はmerkleblockを受け取っては捨て続ける
		<-blockCh
	}
}

func getTxs(conn net.Conn, txCh chan *message.Transaction, doneCh chan struct{}, txs *[]*message.Transaction) {
Loop:
	for {
		select {
		case tx := <-txCh:
			*txs = append(*txs, tx)
		case <-doneCh:
			fmt.Println("tx receive done")
			break Loop
		case <-time.After(time.Second * 30):
			fmt.Println("Fail got transactions")
			break Loop
		}
	}
	for {
		// 以降はtransactionを受け取っては捨て続ける
		<-txCh
	}
}

func dispatch(conn net.Conn, blockCh chan *message.Merkleblock, txCh chan *message.Transaction) {
	var header [common.MessageHeaderLen]byte
	buf := make([]byte, common.MessageHeaderLen)
	t := time.NewTicker(100 * time.Millisecond)
Loop:
	for {
		select {
		case <-t.C:
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println(err.Error())
				break Loop
			}
			if n == common.MessageHeaderLen {
				copy(header[:], buf)
				mh := common.DecodeMessageHeader(header)
				fmt.Printf("Recv: %s %d bytes\n", string(mh.Command[:]), mh.Length)
				msgBytes, err := RecvMessage(conn, mh.Length)
				if err != nil {
					fmt.Println(err.Error())
					break Loop
				}
				if bytes.HasPrefix(mh.Command[:], []byte("inv")) {
					inv, err := message.DecodeInv(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					inventory := []*message.InvVect{}
					for _, invvect := range inv.Inventory {
						if invvect.InvType == message.InvTypeMsgBlock {
							inventory = append(inventory, message.NewInvVect(message.InvTypeMsgFilteredBlock, invvect.Hash))
						} else {
							inventory = append(inventory, invvect)
						}
					}
					getData := message.NewGetData(inventory)
					SendMessage(conn, getData)
				} else if bytes.HasPrefix(mh.Command[:], []byte("merkleblock")) {
					merkleBlock, err := message.DecodeMerkleBlock(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					blockCh <- merkleBlock
				} else if bytes.HasPrefix(mh.Command[:], []byte("tx")) {
					transaction, err := message.DecodeTransaction(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					txID := transaction.ID()
					fmt.Println(hex.EncodeToString(txID[:]))
					txCh <- transaction
				} else if bytes.HasPrefix(mh.Command[:], []byte("reject")) {
					reject, err := message.DecodeReject(msgBytes)
					if err != nil {
						fmt.Println(err.Error())
						break Loop
					}
					fmt.Println(reject.String())
				} else {
					continue
				}
			}
		}
	}
}
