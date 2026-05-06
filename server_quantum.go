package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

type ServerQuantumEncryption struct {
	aesGCM   cipher.AEAD
	chacha   cipher.AEAD
	keyHash  [32]byte
}

func NewServerQuantumEncryption(password, salt string) (*ServerQuantumEncryption, error) {
	key := argon2.IDKey([]byte(password), []byte(salt), 3, 64*1024, 4, 64)
	
	aesKey := key[:32]
	chachaKey := key[32:64]

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	chacha, err := chacha20poly1305.NewX(chachaKey)
	if err != nil {
		return nil, err
	}

	return &ServerQuantumEncryption{
		aesGCM:  aesGCM,
		chacha:  chacha,
		keyHash: sha256.Sum256(key),
	}, nil
}

func (q *ServerQuantumEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	nonce1 := make([]byte, q.aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce1); err != nil {
		return nil, err
	}

	ciphertext1 := q.aesGCM.Seal(nonce1, nonce1, plaintext, nil)

	nonce2 := make([]byte, q.chacha.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce2); err != nil {
		return nil, err
	}

	ciphertext2 := q.chacha.Seal(nonce2, nonce2, ciphertext1, nil)

	return ciphertext2, nil
}

func (q *ServerQuantumEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < q.chacha.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce2 := ciphertext[:q.chacha.NonceSize()]
	ciphertext2 := ciphertext[q.chacha.NonceSize():]

	plaintext1, err := q.chacha.Open(nil, nonce2, ciphertext2, nil)
	if err != nil {
		return nil, err
	}

	if len(plaintext1) < q.aesGCM.NonceSize() {
		return nil, fmt.Errorf("intermediate ciphertext too short")
	}

	nonce1 := plaintext1[:q.aesGCM.NonceSize()]
	ciphertext1 := plaintext1[q.aesGCM.NonceSize():]

	plaintext, err := q.aesGCM.Open(nil, nonce1, ciphertext1, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
