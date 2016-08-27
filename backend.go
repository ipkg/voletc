package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type Backend interface {
	// Get a key value map under a given prefix
	GetMap(string) (map[string][]byte, error)
	// Set key value map under the given prefix
	SetMap(string, map[string][]byte) error
	// Delete all keys under the given prefix
	DeleteMap(string) error
}

func NewBackend(dcfg *DriverConfig) (Backend, error) {
	var (
		be  Backend
		err error
	)

	switch dcfg.BackendType {
	case "consul":
		be, err = NewConsulBackend(dcfg.BackendAddr, dcfg.Prefix)

	default:
		err = fmt.Errorf("backend not supported: %s", dcfg.BackendType)

	}

	// Enable encryption
	if err == nil && len(dcfg.EncryptionKey) > 1 {
		return &BasicEncryptedBackend{be: be, key: []byte(dcfg.EncryptionKey)}, nil
	}

	return be, err
}

type BasicEncryptedBackend struct {
	be Backend

	key []byte
}

// Get a key value map under a given prefix
func (ebe *BasicEncryptedBackend) GetMap(prefix string) (map[string][]byte, error) {
	m, err := ebe.be.GetMap(prefix)
	if err == nil {
		for k, v := range m {
			txt, err := ebe.decrypt(v)
			if err != nil {
				return nil, err
			}

			m[k] = txt
		}
	}

	return m, err
}

// Set key value map under the given prefix
func (ebe *BasicEncryptedBackend) SetMap(prefix string, kmap map[string][]byte) error {
	emap := map[string][]byte{}
	for k, v := range kmap {
		ct, err := ebe.encrypt(v)
		if err != nil {
			return err
		}
		emap[k] = ct
	}

	return ebe.be.SetMap(prefix, emap)
}

// Delete all keys under the given prefix
func (ebe *BasicEncryptedBackend) DeleteMap(prefix string) error {
	return ebe.be.DeleteMap(prefix)
}

func (ebe *BasicEncryptedBackend) encrypt(text []byte) ([]byte, error) {
	return encrypt(ebe.key, text)
}

func (ebe *BasicEncryptedBackend) decrypt(ciphertext []byte) ([]byte, error) {
	return decrypt(ebe.key, ciphertext)
}

func encrypt(key, text []byte) (ciphertext []byte, err error) {

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}

	ciphertext = make([]byte, aes.BlockSize+len(string(text)))

	// iv =  initialization vector
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return
}

func decrypt(key, ciphertext []byte) (plaintext []byte, err error) {

	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(ciphertext) < aes.BlockSize {
		err = fmt.Errorf("ciphertext too short")
		return
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	plaintext = ciphertext

	return
}
