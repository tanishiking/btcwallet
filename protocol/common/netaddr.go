package common

import (
	"encoding/binary"
)

// NetAddr means network address.
type NetAddr struct {
	Services uint64   // versionのserviceと同様
	IP       [16]byte // IPアドレス
	Port     uint16   // ポート番号 big endian
}

// DecodeNetAddr decodes byte array to NetAddr
func DecodeNetAddr(b [26]byte) *NetAddr {
	var ip [16]byte
	copy(ip[:], b[8:24])
	return &NetAddr{
		Services: binary.LittleEndian.Uint64(b[0:8]),
		IP:       ip,
		Port:     binary.BigEndian.Uint16(b[24:26]),
	}
}

// Encode encode NetAddr to byte array.
func (addr *NetAddr) Encode() [26]byte {
	var b [26]byte
	binary.LittleEndian.PutUint64(b[0:8], addr.Services)
	copy(b[8:24], addr.IP[:])
	binary.BigEndian.PutUint16(b[24:26], addr.Port)
	return b
}
