package auth

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

func UpenSealedBox(privateKey [32]byte, encryptedB64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return nil, err
	}
	if len(data) < 32 {
		return nil, fmt.Errorf("data too short")
	}
	var ephPub [32]byte
	copy(ephPub[:], data[:32])
	ciphertext := data[32:]
	var myPub [32]byte
	curve25519.ScalarBaseMult(&myPub, &privateKey)
	nonce, err := sealNonce(ephPub[:], myPub[:])
	if err != nil {
		return nil, err
	}
	plaintext, ok := box.Open(nil, ciphertext, &nonce, &ephPub, &privateKey)
	if !ok {
		return nil, fmt.Errorf("failed to open sealed box")
	}
	return plaintext, nil
}

func sealNonce(sphPub, recipientPub []byte) ([24]byte, error) {
	var nonce [24]byte
	h, err := blake2b.New(24, nil)
	if err != nil {
		return nonce, err
	}
	h.Write(sphPub)
	h.Write(recipientPub)
	copy(nonce[:], h.Sum(nil))
	return nonce, nil
}
