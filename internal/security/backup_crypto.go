package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func GenerateEncryptedBackup(passphrase string, outputPath string) (string, error) {
	if passphrase == "" {
		passphrase = "ClawRT-Default-Key-2026"
	}
	if outputPath == "" {
		outputPath = "/tmp/backup-clawrt-encrypted.enc"
	}

	rawBackup := "/tmp/backup-raw-temp.tar.gz"
	defer os.Remove(rawBackup)

	// 1. Generate raw sysupgrade backup
	cmd := exec.Command("sysupgrade", "-b", rawBackup)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("error en sysupgrade backup: %v (%s)", err, string(out))
	}

	// 2. Read raw backup bytes
	rawBytes, err := os.ReadFile(rawBackup)
	if err != nil {
		return "", fmt.Errorf("error al leer archivo de respaldo: %w", err)
	}

	// 3. Derive 32-byte key from passphrase
	key := sha256.Sum256([]byte(passphrase))

	// 4. Encrypt with AES-256-GCM
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, rawBytes, nil)

	// 5. Write encrypted file
	if err := os.WriteFile(outputPath, ciphertext, 0600); err != nil {
		return "", fmt.Errorf("error al guardar archivo cifrado: %w", err)
	}

	return fmt.Sprintf("🔐 Respaldo cifrado con éxito (AES-256-GCM) en %s (%d bytes).", outputPath, len(ciphertext)), nil
}
