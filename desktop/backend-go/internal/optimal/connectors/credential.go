package connectors

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// Envelope format, identical byte-for-byte to the Elixir source:
//
//	<<version::8, iv::96, tag::128, ciphertext::binary>>
//
// stored as Base64 in the `config.credentials_ciphertext` JSON field.
// version = 1 today; any change bumps it and triggers rotation.
const (
	credentialVersion = 1
	credentialIVSize  = 12
	credentialTagSize = 16
	credentialKeySize = 32 // AES-256 requires a 32-byte key
)

// Credential errors exposed to callers.
var (
	ErrCredentialKeyMissing  = errors.New("connectors/credential: CONNECTOR_KEY env missing or malformed")
	ErrCredentialBadEnvelope = errors.New("connectors/credential: bad envelope")
	ErrCredentialVersion     = errors.New("connectors/credential: unsupported envelope version")
	ErrCredentialAuthFailed  = errors.New("connectors/credential: auth tag verification failed")
)

// EncryptCredentials serializes creds to JSON and AES-256-GCM envelope-encrypts
// it with the master key from the CONNECTOR_KEY env var. Returns a Base64 string.
func EncryptCredentials(creds map[string]any) (string, error) {
	key, err := masterKey()
	if err != nil {
		return "", err
	}
	plaintext, err := json.Marshal(creds)
	if err != nil {
		return "", fmt.Errorf("connectors/credential: marshal: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("connectors/credential: cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, credentialIVSize)
	if err != nil {
		return "", fmt.Errorf("connectors/credential: gcm: %w", err)
	}

	iv := make([]byte, credentialIVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("connectors/credential: iv: %w", err)
	}

	// Seal appends the 16-byte tag to the ciphertext. We split them so the
	// on-wire layout matches the Elixir <<iv, tag, ciphertext>> ordering.
	sealed := gcm.Seal(nil, iv, plaintext, nil)
	ciphertext := sealed[:len(sealed)-credentialTagSize]
	tag := sealed[len(sealed)-credentialTagSize:]

	envelope := make([]byte, 0, 1+credentialIVSize+credentialTagSize+len(ciphertext))
	envelope = append(envelope, byte(credentialVersion))
	envelope = append(envelope, iv...)
	envelope = append(envelope, tag...)
	envelope = append(envelope, ciphertext...)

	return base64.StdEncoding.EncodeToString(envelope), nil
}

// DecryptCredentials inverts EncryptCredentials. Errors are typed so callers
// can distinguish "key missing" from "auth failed" from "bad envelope".
func DecryptCredentials(envelope string) (map[string]any, error) {
	key, err := masterKey()
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(envelope)
	if err != nil {
		return nil, ErrCredentialBadEnvelope
	}
	if len(raw) < 1+credentialIVSize+credentialTagSize {
		return nil, ErrCredentialBadEnvelope
	}
	version := raw[0]
	if version != credentialVersion {
		return nil, ErrCredentialVersion
	}
	iv := raw[1 : 1+credentialIVSize]
	tag := raw[1+credentialIVSize : 1+credentialIVSize+credentialTagSize]
	ciphertext := raw[1+credentialIVSize+credentialTagSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("connectors/credential: cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, credentialIVSize)
	if err != nil {
		return nil, fmt.Errorf("connectors/credential: gcm: %w", err)
	}

	// Go's AEAD API expects ciphertext||tag; we stored them apart so rejoin.
	sealed := append(append([]byte{}, ciphertext...), tag...)
	plaintext, err := gcm.Open(nil, iv, sealed, nil)
	if err != nil {
		return nil, ErrCredentialAuthFailed
	}

	var out map[string]any
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, fmt.Errorf("connectors/credential: unmarshal: %w", err)
	}
	return out, nil
}

// CredentialsReady mirrors Credential.ready?/0 — handy for startup health checks.
func CredentialsReady() bool {
	_, err := masterKey()
	return err == nil
}

func masterKey() ([]byte, error) {
	hex64 := os.Getenv("CONNECTOR_KEY")
	if len(hex64) != 64 {
		return nil, ErrCredentialKeyMissing
	}
	key, err := hex.DecodeString(hex64)
	if err != nil || len(key) != credentialKeySize {
		return nil, ErrCredentialKeyMissing
	}
	return key, nil
}
