package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
)

// ── Secrets ──────────────────────────────────────────────────────────────────

type Secret struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Created     string `json:"created"`
	LastUpdated string `json:"lastUpdated"`
}

type SecretPage struct {
	Content       []Secret `json:"content"`
	TotalElements int      `json:"totalElements"`
	TotalPages    int      `json:"totalPages"`
	Number        int      `json:"number"`
	Last          bool     `json:"last"`
}

type createSecretAPIRequest struct {
	Name           string `json:"name"`
	EncryptedValue string `json:"encryptedValue"`
	EncryptionKey  string `json:"encryptionKey"`
}

type CreateSecretResponse struct {
	Secret Secret `json:"secret"`
}

type updateSecretAPIRequest struct {
	EncryptedValue string `json:"encryptedValue"`
	EncryptionKey  string `json:"encryptionKey"`
}

type UpdateSecretResponse struct {
	Secret Secret `json:"secret"`
}

type jwk struct {
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type publicKeyResponse struct {
	EncryptionKey jwk `json:"encryptionKey"`
}

func (c *Client) ListSecrets(page, size int) (*SecretPage, error) {
	params := url.Values{}
	if size <= 0 {
		size = 100
	}
	params.Set("size", strconv.Itoa(size))
	params.Set("page", strconv.Itoa(page))
	resp, err := c.get("/api/v1/secrets", params)
	if err != nil {
		return nil, err
	}
	var result SecretPage
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetSecret(secretID string) (*Secret, error) {
	resp, err := c.get("/api/v1/secrets/"+secretID, nil)
	if err != nil {
		return nil, err
	}
	var result Secret
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSecretsPublicKey fetches the RSA public key used to encrypt secret values.
func (c *Client) GetSecretsPublicKey() (*rsa.PublicKey, error) {
	resp, err := c.get("/api/v1/secretsPublicKey", nil)
	if err != nil {
		return nil, err
	}
	var result publicKeyResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return parseJWK(result.EncryptionKey)
}

// CreateSecret encrypts the value client-side and sends it to the API.
func (c *Client) CreateSecret(name, plainValue string) (*Secret, error) {
	pubKey, err := c.GetSecretsPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch encryption key: %w", err)
	}
	encryptedValue, encryptionKey, err := encryptSecretValue(plainValue, pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}
	resp, err := c.post("/api/v1/secrets", createSecretAPIRequest{
		Name:           name,
		EncryptedValue: encryptedValue,
		EncryptionKey:  encryptionKey,
	})
	if err != nil {
		return nil, err
	}
	var result CreateSecretResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result.Secret, nil
}

// UpdateSecret encrypts the new value client-side and sends it to the API.
func (c *Client) UpdateSecret(secretID, plainValue string) (*Secret, error) {
	pubKey, err := c.GetSecretsPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch encryption key: %w", err)
	}
	encryptedValue, encryptionKey, err := encryptSecretValue(plainValue, pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}
	resp, err := c.post("/api/v1/secrets/"+secretID, updateSecretAPIRequest{
		EncryptedValue: encryptedValue,
		EncryptionKey:  encryptionKey,
	})
	if err != nil {
		return nil, err
	}
	var result UpdateSecretResponse
	if err := decode(resp, &result); err != nil {
		return nil, err
	}
	return &result.Secret, nil
}

func (c *Client) DeleteSecret(secretID string) error {
	resp, err := c.delete("/api/v1/secrets/" + secretID)
	if err != nil {
		return err
	}
	return decode(resp, &struct{}{})
}

// ── Encryption helpers ───────────────────────────────────────────────────────

// encryptSecretValue implements the two-layer encryption scheme:
//  1. Generate random AES-256-GCM key (32 bytes) + IV (12 bytes)
//  2. Encrypt plaintext with AES-256-GCM → base64 → prefix "encrypted_" → encryptedValue
//  3. Concatenate "{base64_key}:::{base64_iv}" → RSA-OAEP(SHA-256) → base64 → encryptionKey
func encryptSecretValue(plaintext string, pubKey *rsa.PublicKey) (encryptedValue, encryptionKey string, err error) {
	// Step 1: generate AES key and IV
	aesKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(aesKey); err != nil {
		return "", "", fmt.Errorf("generate AES key: %w", err)
	}
	iv := make([]byte, 12) // GCM standard nonce
	if _, err := rand.Read(iv); err != nil {
		return "", "", fmt.Errorf("generate IV: %w", err)
	}

	// Step 2: encrypt with AES-256-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", "", fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("create GCM: %w", err)
	}
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	encryptedValue = "encrypted_" + base64.StdEncoding.EncodeToString(ciphertext)

	// Step 3: encrypt the AES key+IV with RSA-OAEP
	keyMaterial := base64.StdEncoding.EncodeToString(aesKey) + ":::" + base64.StdEncoding.EncodeToString(iv)
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(keyMaterial), nil)
	if err != nil {
		return "", "", fmt.Errorf("RSA encrypt: %w", err)
	}
	encryptionKey = base64.StdEncoding.EncodeToString(rsaCiphertext)

	return encryptedValue, encryptionKey, nil
}

// parseJWK converts a JWK RSA key (with base64url-encoded n and e) to *rsa.PublicKey.
func parseJWK(k jwk) (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", k.Kty)
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())
	return &rsa.PublicKey{N: n, E: e}, nil
}
