package key

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"

	secp256k1 "github.com/toxeus/go-secp256k1"

	"github.com/tanishiking/btcwallet/util"
)

const (
	size              = 32
	secretKeyFilePath = "secretkey"
)

// ReadOrGeneratePrivateKey read or generate private key.
func ReadOrGeneratePrivateKey() ([]byte, error) {
	_, err := os.Stat(secretKeyFilePath)
	if err == nil {
		data, err := ioutil.ReadFile(secretKeyFilePath)
		if err != nil {
			return []byte{}, err
		}
		return DecodeWIF(string(data)), nil
	}
	priv := GeneratePrivateKey()
	wif := EncodeWIF(priv)
	err = ioutil.WriteFile(secretKeyFilePath, []byte(wif), 0666)
	if err != nil {
		return []byte{}, err
	}
	return priv, nil
}

// GeneratePrivateKey generate new private key.
func GeneratePrivateKey() []byte {
	var privateKeyBytes32 [32]byte
	secp256k1.Start()
Loop:
	for {
		for i := 0; i < size; i++ {
			//This is not "cryptographically random"
			privateKeyBytes32[i] = byte(util.RandInt(0, math.MaxUint8))
		}
		ok := secp256k1.Seckey_verify(privateKeyBytes32)
		if ok {
			break Loop
		}
	}
	secp256k1.Stop()
	return privateKeyBytes32[:]
}

// GeneratePubKey generate new public key from private key.
func GeneratePubKey(privateKeyBytes []byte) ([]byte, error) {
	var privateKeyBytes32 [size]byte
	for i := 0; i < size; i++ {
		privateKeyBytes32[i] = privateKeyBytes[i]
	}
	secp256k1.Start()
	publicKeyBytes, success := secp256k1.Pubkey_create(privateKeyBytes32, false)
	secp256k1.Stop()

	if !success {
		return []byte{}, fmt.Errorf("Failed to generate public key")
	}
	return publicKeyBytes, nil
}
