package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/tanishiking/btcwallet/protocol/common"
	"github.com/tanishiking/btcwallet/protocol/message"
	"github.com/tanishiking/btcwallet/util"
)

// CreateMessageHeader create messageheader from message.
func CreateMessageHeader(msg Message) *common.MessageHeader {
	var (
		commandNameBytes [12]byte
		checksum         [4]byte
	)
	hashedMsg := util.Hash256(msg.Encode())
	copy(commandNameBytes[:], []byte(msg.CommandName()))
	copy(checksum[:], hashedMsg[0:4])
	return &common.MessageHeader{
		Magic:    binary.LittleEndian.Uint32([]byte{0x0B, 0x11, 0x09, 0x07}), // testnet magic value
		Command:  commandNameBytes,
		Length:   uint32(len(msg.Encode())),
		Checksum: checksum,
	}
}

// RecvMessage receive sized message from the connection .
func RecvMessage(conn net.Conn, size uint32) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	_, err := conn.Read(buf)
	if err != nil {
		return []byte{}, err
	}
	return buf, nil
}

// SendMessage send the message to remote peer via the connection.
func SendMessage(conn net.Conn, msg Message) error {
	header := CreateMessageHeader(msg)
	payload := msg.Encode()
	// size := len(payload)
	data := bytes.Join([][]byte{header.Encode(), payload}, []byte{})
	_, err := conn.Write(data)
	if err != nil {
		fmt.Printf("Message send failed %v \n", msg)
		return err
	}
	fmt.Printf("Send %s: %d bytes\n", msg.CommandName(), len(payload))
	return nil
}

// WithBitcoinConnection connect to fallback node in testnet and then
// do the received function using the connection with the node.
func WithBitcoinConnection(fn func(net.Conn, *message.Version)) {
	conn, err := net.Dial("tcp", "testnet-seed.bitcoin.jonasschnelli.ch:18333")
	if err != nil {
		fmt.Println("Failed to connect to peer: ", err.Error())
		return
	}
	fmt.Printf("Connected: %#v \n", conn.RemoteAddr().String())

	verackCh := make(chan *message.Verack)
	versionCh := make(chan *message.Version)
	errCh := make(chan error)

	go func(conn net.Conn, verackCh chan *message.Verack, versioinCh chan *message.Version, errCh chan error) {
		var header [24]byte
		recvVerack := false
		recvVersion := false
		buf := make([]byte, 24)
	Loop:
		for {
			if recvVerack && recvVersion {
				break Loop
			}
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				errCh <- err
				continue
			}
			if n == 24 {
				copy(header[:], buf)
				mh := common.DecodeMessageHeader(header)
				command := string(mh.Command[:])
				fmt.Printf("Recv: %s %d\n", command, mh.Length)
				payload, err := RecvMessage(conn, mh.Length)
				if err != nil {
					errCh <- err
				}
				if bytes.HasPrefix(mh.Command[:], []byte("verack")) {
					verackCh <- &message.Verack{}
					recvVerack = true
				} else if bytes.HasPrefix(mh.Command[:], []byte("version")) {
					v, err := message.DecodeVersion(payload)
					if err != nil {
						errCh <- err
						continue
					}
					versionCh <- v
					recvVersion = true
				}
			}
		}
	}(conn, verackCh, versionCh, errCh)

	addrFrom := &common.NetAddr{
		Services: uint64(1),
		IP: [16]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x7F, 0x00, 0x00, 0x01,
		}, // 127.0.0.1 https://en.bitcoin.it/wiki/Protocol_documentation#Network_address
		Port: 8333,
	}
	v := &message.Version{
		Version:     uint32(70015),
		Services:    uint64(1),
		Timestamp:   uint64(time.Now().Unix()),
		AddrRecv:    addrFrom, // 適当、remote peer はconnection時にこのフィールド見てない？
		AddrFrom:    addrFrom,
		Nonce:       uint64(0), //  connection時のnonceこれでいいのか
		UserAgent:   common.NewVarStr([]byte("")),
		StartHeight: uint32(0),
		Relay:       false,
	}
	err = SendMessage(conn, v)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	recvVerack := false
	recvVersion := false
	var receivedVersion *message.Version
	// Wait for the verack message or timeout in case of failure.
Loop:
	for {
		if recvVerack && recvVersion {
			fn(conn, receivedVersion)
			break Loop
		}
		select {
		case <-verackCh:
			recvVerack = true
		case receivedVersion = <-versionCh:
			recvVersion = true
			SendMessage(conn, &message.Verack{})
		case err := <-errCh:
			fmt.Println(err.Error())
			return
		case <-time.After(time.Second * 5):
			fmt.Printf("Peer Connection: verack/version timeout")
			return
		}
	}
}
