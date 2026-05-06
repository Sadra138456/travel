package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type ClientQuantumEncryption struct {
	key []byte
	gcm cipher.AEAD
}

func NewClientQuantumEncryption(password, salt string) (*ClientQuantumEncryption, error) {
	key := pbkdf2.Key([]byte(password), []byte(salt), 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &ClientQuantumEncryption{
		key: key,
		gcm: gcm,
	}, nil
}

func (e *ClientQuantumEncryption) Encrypt(plaintext []byte) []byte {
	nonce := make([]byte, e.gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return e.gcm.Seal(nonce, nonce, plaintext, nil)
}

func (e *ClientQuantumEncryption) Decrypt(ciphertext []byte) []byte {
	nonceSize := e.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil
	}

	return plaintext
}
