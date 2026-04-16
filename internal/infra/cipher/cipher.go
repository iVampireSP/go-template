package cipher

import (
	"crypto/aes"
	gcmcipher "crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/iVampireSP/go-template/internal/infra/keystore"
)

// Cipher AES-256-GCM 对称加密，密钥材料来自 keystore RSA 私钥
type Cipher struct {
	key []byte // 32 bytes for AES-256
}

// NewCipher 从 keystore 获取 RSA 密钥，派生 AES-256 密钥
func NewCipher(ks *keystore.KeyStore) (*Cipher, error) {
	rsaKey, err := ks.RSA("rsa")
	if err != nil {
		return nil, fmt.Errorf("cipher: RSA key 'rsa' required: %w", err)
	}
	hash := sha256.Sum256(rsaKey.PrivateKeyPEM)
	return &Cipher{
		key: hash[:],
	}, nil
}

// Encrypt 加密数据，返回 base64 编码的密文（nonce + ciphertext）
func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := gcmcipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文
func (c *Cipher) Decrypt(encoded string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := gcmcipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
