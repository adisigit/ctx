package auth

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"

	"golang.org/x/crypto/curve25519"
)

type KeyPair struct {
	PublicKey  [32]byte
	PrivateKey [32]byte
}

func GenerateKeyPair() (*KeyPair, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return nil, err
	}
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64
	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &priv)
	return &KeyPair{PublicKey: pub, PrivateKey: priv}, nil
}

func (k *KeyPair) PublicKeyB64() string {
	return base64.StdEncoding.EncodeToString(k.PublicKey[:])
}

func (k *KeyPair) SaveToDisk(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	privB64 := base64.StdEncoding.EncodeToString(k.PrivateKey[:])
	return os.WriteFile(filepath.Join(dir, "identity.key"), []byte(privB64), 0600)
}

func LoadKeyPair(dir string) (*KeyPair, error) {
	privB64, err := os.ReadFile(filepath.Join(dir, "identity.key"))
	if err != nil {
		return nil, err
	}
	var priv [32]byte
	if _, err := base64.StdEncoding.Decode(priv[:], privB64); err != nil {
		return nil, err
	}
	kp := &KeyPair{}
	copy(kp.PrivateKey[:], priv[:])
	curve25519.ScalarBaseMult(&kp.PublicKey, &kp.PrivateKey)
	return kp, nil
}
