package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// backupSecret is the login material for an sftp, webdav or cloud-drive destination.
type backupSecret struct {
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	// Passphrase unlocks a passphrase-protected PrivateKey.
	Passphrase string `json:"passphrase,omitempty"`
	// RefreshToken is the OAuth refresh token of a cloud-drive destination.
	RefreshToken string `json:"refreshToken,omitempty"`

	// accessToken and accessExpiry seed a cloud-drive client that is only
	// being tested. They are never serialized or stored.
	accessToken  string
	accessExpiry time.Time
}

func (s backupSecret) empty() bool {
	return s.Password == "" && s.PrivateKey == "" && s.RefreshToken == ""
}

// secretBox seals destination credentials with AES-256-GCM. The key is
// derived from HBOX_BACKUP_ENCRYPTION_KEY; the additional data binds a blob to
// its destination so a sealed secret copied onto another row cannot be opened.
type secretBox struct {
	aead cipher.AEAD
}

func newSecretBox(key string) (*secretBox, error) {
	if key == "" {
		return nil, nil //nolint:nilnil // no key configured: credentials-based destinations are disabled
	}
	sum := sha256.Sum256([]byte("homebox-backup-secrets-v1:" + key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &secretBox{aead: aead}, nil
}

func (b *secretBox) seal(aad string, v backupSecret) (string, error) {
	plain, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := b.aead.Seal(nonce, nonce, plain, []byte(aad))
	return base64.StdEncoding.EncodeToString(out), nil
}

func (b *secretBox) open(aad, blob string) (backupSecret, error) {
	var v backupSecret
	raw, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return v, errors.New("stored credentials are unreadable")
	}
	ns := b.aead.NonceSize()
	if len(raw) < ns {
		return v, errors.New("stored credentials are unreadable")
	}
	plain, err := b.aead.Open(nil, raw[:ns], raw[ns:], []byte(aad))
	if err != nil {
		return v, errors.New("stored credentials cannot be decrypted; the encryption key may have changed")
	}
	if err := json.Unmarshal(plain, &v); err != nil {
		return v, fmt.Errorf("stored credentials are corrupt: %w", err)
	}
	return v, nil
}
