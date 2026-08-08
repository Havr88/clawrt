package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

// CryptoEngine maneja las operaciones criptográficas ligeras para ClawRT
type CryptoEngine struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

func NewCryptoEngine() (*CryptoEngine, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("error al generar llaves Ed25519: %w", err)
	}
	return &CryptoEngine{
		publicKey:  pub,
		privateKey: priv,
	}, nil
}

// DeriveKeyArgon2id deriva una clave simétrica de 32 bytes optimizada para routers con 64MB RAM
func DeriveKeyArgon2id(passphrase []byte, salt []byte) []byte {
	// Parámetros adaptados a routers MIPS: 1 iteración, 8 MB RAM, 1 hilo
	return argon2.IDKey(passphrase, salt, 1, 8*1024, 1, 32)
}

// EncryptDataChacha/Ascon encrypts sensitive state with AEAD ChaCha20-Poly1305 / Lightweight AEAD
func EncryptSensitiveData(plaintext []byte, key []byte) (string, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSensitiveData decrypts sensitive data
func DecryptSensitiveData(encryptedB64 string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	if len(data) < aead.NonceSize() {
		return nil, fmt.Errorf("datos cifrados inválidos o demasiado cortos")
	}

	nonce, ciphertext := data[:aead.NonceSize()], data[aead.NonceSize():]
	return aead.Open(nil, nonce, ciphertext, nil)
}

// SignMessage firma digitalmente una trama de datos con Ed25519
func (ce *CryptoEngine) SignMessage(message []byte) string {
	sig := ed25519.Sign(ce.privateKey, message)
	return base64.StdEncoding.EncodeToString(sig)
}

// VerifySignature verifica la firma Ed25519 de un mensaje
func VerifySignature(pubKey ed25519.PublicKey, message []byte, sigB64 string) bool {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false
	}
	return ed25519.Verify(pubKey, message, sig)
}

// ComputeHash SHA-256 ligero para checksums de configuración UCI
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
