package message

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/tanishiking/btcwallet/protocol/common"
)

// Version means version message.
type Version struct {
	Version     uint32          // ノードで使われているプロトコルのバージョン番号
	Services    uint64          //  ノードがサポートしている機能を表すフラグ
	Timestamp   uint64          //  UNIXタイムスタンプ（秒）
	AddrRecv    *common.NetAddr //  受け手のネットワーク・アドレス
	AddrFrom    *common.NetAddr //  送り手のネットワーク・アドレス
	Nonce       uint64          //  コネクションを特定するために使用されるランダムな値
	UserAgent   *common.VarStr  //  ユーザーエージェント。see. BIP14
	StartHeight uint32          //  ノードが持っているブロックの高さ
	Relay       bool            //  接続先ノードが受信したトランザクションを送っていいかどうか
}

// DecodeVersion decode byte slice to version.
func DecodeVersion(b []byte) (*Version, error) {
	length := len(b)
	if length <= 80 {
		return nil, fmt.Errorf("Invalid version message: %#v", b)
	}
	var addrRecvArr [26]byte
	var addrFromArr [26]byte
	versionByte := binary.LittleEndian.Uint32(b[0:4])
	services := binary.LittleEndian.Uint64(b[4:12])
	timestamp := binary.LittleEndian.Uint64(b[12:20])

	copy(addrRecvArr[:], b[20:46])
	addrRecv := common.DecodeNetAddr(addrRecvArr)
	fmt.Println("addrRecv ip: ", addrRecv.IP, addrRecv.Port)

	copy(addrFromArr[:], b[46:72])
	addrFrom := common.DecodeNetAddr(addrFromArr)
	fmt.Println("addrFrom ip: ", addrFrom.IP, addrFrom.Port)

	nonce := binary.LittleEndian.Uint64(b[72:80])

	// userAgent の読み取り
	userAgent, err := common.DecodeVarStr(b[80:])
	if err != nil {
		return nil, err
	}
	varstrLen := len(userAgent.Encode())
	fmt.Println("UserAgent: ", string(userAgent.Data))

	if length < 85+varstrLen {
		return nil, fmt.Errorf("Invalid version message: %#v", b)
	}

	startHeight := binary.LittleEndian.Uint32(b[80+varstrLen : 84+varstrLen])
	var relay bool
	if b[84+varstrLen] > 0x00 {
		relay = true
	} else {
		relay = false
	}
	return &Version{
		Version:     versionByte,
		Services:    services,
		Timestamp:   timestamp,
		AddrRecv:    addrRecv,
		AddrFrom:    addrFrom,
		Nonce:       nonce,
		UserAgent:   userAgent,
		StartHeight: startHeight,
		Relay:       relay,
	}, nil
}

// CommandName return version.
func (v *Version) CommandName() string {
	return "version"
}

// Encode encode version to byte slice.
func (v *Version) Encode() []byte {
	var (
		version     [4]byte
		services    [8]byte
		timestamp   [8]byte
		addrRecv    [26]byte
		addrFrom    [26]byte
		nonce       [8]byte
		userAgent   []byte
		startHeight [4]byte
		relay       [1]byte
	)

	binary.LittleEndian.PutUint32(version[:4], v.Version)
	binary.LittleEndian.PutUint64(services[:8], v.Services)
	binary.LittleEndian.PutUint64(timestamp[:8], v.Timestamp)
	addrRecv = v.AddrRecv.Encode()
	addrFrom = v.AddrFrom.Encode()
	binary.LittleEndian.PutUint64(nonce[:8], v.Nonce)
	userAgent = v.UserAgent.Encode()
	binary.LittleEndian.PutUint32(startHeight[:4], v.StartHeight)
	if v.Relay {
		relay = [1]byte{0x01}
	} else {
		relay = [1]byte{0x00}
	}
	return bytes.Join(
		[][]byte{
			version[:],
			services[:],
			timestamp[:],
			addrRecv[:],
			addrFrom[:],
			nonce[:],
			userAgent[:],
			startHeight[:],
			relay[:],
		},
		[]byte{},
	)
}
