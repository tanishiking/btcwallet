package key

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mr-tron/base58/base58"

	"github.com/tanishiking/btcwallet/util"
)

// EncodeWIF encodes private key to Wallet Import Format
//
// refer: https://en.bitcoin.it/wiki/Wallet_import_format
func EncodeWIF(privateKeyBytes []byte) string {
	// 1. 先頭にネットワークを表す1byteのprefixをつける
	bs := bytes.Join([][]byte{
		[]byte{0xEF}, // This means that, this address is for testnet.
		privateKeyBytes,
	},
		[]byte{},
	)
	// 2. (1)にutil.Hash256を適用したものの先頭4バイトの文字をとる
	checksum := util.Hash256(bs)[:4]

	// 3 (1) と (2) を結合して base58 encode
	return base58.Encode(bytes.Join([][]byte{bs, checksum}, []byte{}))
}

// DecodeWIF decodes wallet import format byte string to private key
func DecodeWIF(wif string) []byte {
	decoded, err := base58.Decode(wif)
	if err != nil {
		// TODO: error handling ...
		fmt.Println(err.Error())
		os.Exit(1)
	}

	bs := decoded[:len(decoded)-4]
	// checksum := decoded[len(decoded)-4:]
	return bs[1:]
}

// EncodeBitcoinAddr encode public key to bitcoin address.
//
// refer: https://en.bitcoin.it/w/index.php?title=Technical_background_of_version_1_Bitcoin_addresses
func EncodeBitcoinAddr(publicKeyBytes []byte) string {
	bs := bytes.Join([][]byte{
		[]byte{0x6F}, // This means that, this address is for testnet.
		util.Hash160(publicKeyBytes),
	},
		[]byte{})
	checksum := util.Hash256(bs)[:4]

	return base58.Encode(bytes.Join([][]byte{bs, checksum}, []byte{}))
}

// DecodeBitcoinAddr decode bitcoin address to public key.
//
// refer: https://en.bitcoin.it/w/index.php?title=Technical_background_of_version_1_Bitcoin_addresses
func DecodeBitcoinAddr(addr string) ([]byte, error) {
	decoded, err := base58.Decode(addr)
	if err != nil {
		return nil, err
	}
	bs := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	if !bytes.Equal(util.Hash256(bs)[:4], checksum) {
		return nil, fmt.Errorf("Encode failed: invalid bitcoin address: %s", addr)
	}
	return bs[1:], nil
}
