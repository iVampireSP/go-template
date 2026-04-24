package keystore

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/iVampireSP/go-template/pkg/foundation/config"
)

// decodeKeyData 自动检测并解码密钥数据
// 如果输入是 base64 编码则解码，如果是原始 PEM 格式则直接返回
func decodeKeyData(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, fmt.Errorf("empty key data")
	}

	// PEM 格式以 "-----BEGIN " 开头
	if strings.HasPrefix(data, "-----BEGIN ") {
		return []byte(data), nil
	}

	// 尝试 base64 解码
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key data: not valid PEM or base64 format: %w", err)
	}

	return decoded, nil
}

// KeyStore 密钥存储，支持多个命名密钥
type KeyStore struct {
	mu    sync.RWMutex
	rsa   map[string]*RSAKeyPair
	ecdsa map[string]*ECDSAKeyPair
}

// NewKeyStore 从配置创建密钥存储
func NewKeyStore() (*KeyStore, error) {
	ks := &KeyStore{
		rsa:   make(map[string]*RSAKeyPair),
		ecdsa: make(map[string]*ECDSAKeyPair),
	}

	// 加载配置中定义的所有密钥
	keysConfig := config.Map("keystore.keys")
	for name, v := range keysConfig {
		keyConfig, ok := v.(map[string]any)
		if !ok {
			continue
		}

		keyType, _ := keyConfig["type"].(string)
		if keyType == "" {
			return nil, fmt.Errorf("key '%s' missing required 'type' field", name)
		}

		privateKeyData, _ := keyConfig["private_key"].(string)
		publicKeyData, _ := keyConfig["public_key"].(string)

		if privateKeyData == "" || publicKeyData == "" {
			continue // 跳过未配置的密钥
		}

		switch keyType {
		case "rsa":
			if err := ks.LoadRSA(name, privateKeyData, publicKeyData); err != nil {
				return nil, fmt.Errorf("failed to load RSA key '%s': %w", name, err)
			}
		case "ecdsa":
			if err := ks.LoadECDSA(name, privateKeyData, publicKeyData); err != nil {
				return nil, fmt.Errorf("failed to load ECDSA key '%s': %w", name, err)
			}
		default:
			return nil, fmt.Errorf("unsupported key type '%s' for key '%s'", keyType, name)
		}
	}

	return ks, nil
}

// LoadRSA 加载 RSA 密钥，自动检测 base64 或原始 PEM 格式
func (ks *KeyStore) LoadRSA(name, privateKeyData, publicKeyData string) error {
	privateKeyPEM, err := decodeKeyData(privateKeyData)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyPEM, err := decodeKeyData(publicKeyData)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	keyPair, err := ParseRSAKeyPair(privateKeyPEM, publicKeyPEM)
	if err != nil {
		return err
	}

	ks.mu.Lock()
	ks.rsa[name] = keyPair
	ks.mu.Unlock()

	return nil
}

// LoadECDSA 加载 ECDSA 密钥，自动检测 base64 或原始 PEM 格式
func (ks *KeyStore) LoadECDSA(name, privateKeyData, publicKeyData string) error {
	privateKeyPEM, err := decodeKeyData(privateKeyData)
	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyPEM, err := decodeKeyData(publicKeyData)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	keyPair, err := ParseECDSAKeyPair(privateKeyPEM, publicKeyPEM)
	if err != nil {
		return err
	}

	ks.mu.Lock()
	ks.ecdsa[name] = keyPair
	ks.mu.Unlock()

	return nil
}

// LoadRSAFromBase64 从 base64 编码加载 RSA 密钥
// Deprecated: 使用 LoadRSA 代替，它会自动检测格式
func (ks *KeyStore) LoadRSAFromBase64(name, privateKeyB64, publicKeyB64 string) error {
	privateKeyPEM, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return fmt.Errorf("failed to decode private key from base64: %w", err)
	}

	publicKeyPEM, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return fmt.Errorf("failed to decode public key from base64: %w", err)
	}

	keyPair, err := ParseRSAKeyPair(privateKeyPEM, publicKeyPEM)
	if err != nil {
		return err
	}

	ks.mu.Lock()
	ks.rsa[name] = keyPair
	ks.mu.Unlock()

	return nil
}

// LoadECDSAFromBase64 从 base64 编码加载 ECDSA 密钥
// Deprecated: 使用 LoadECDSA 代替，它会自动检测格式
func (ks *KeyStore) LoadECDSAFromBase64(name, privateKeyB64, publicKeyB64 string) error {
	privateKeyPEM, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return fmt.Errorf("failed to decode private key from base64: %w", err)
	}

	publicKeyPEM, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return fmt.Errorf("failed to decode public key from base64: %w", err)
	}

	keyPair, err := ParseECDSAKeyPair(privateKeyPEM, publicKeyPEM)
	if err != nil {
		return err
	}

	ks.mu.Lock()
	ks.ecdsa[name] = keyPair
	ks.mu.Unlock()

	return nil
}

// SetRSA 直接设置 RSA 密钥对
func (ks *KeyStore) SetRSA(name string, keyPair *RSAKeyPair) {
	ks.mu.Lock()
	ks.rsa[name] = keyPair
	ks.mu.Unlock()
}

// SetECDSA 直接设置 ECDSA 密钥对
func (ks *KeyStore) SetECDSA(name string, keyPair *ECDSAKeyPair) {
	ks.mu.Lock()
	ks.ecdsa[name] = keyPair
	ks.mu.Unlock()
}

// RSA 获取指定名称的 RSA 密钥对
func (ks *KeyStore) RSA(name string) (*RSAKeyPair, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	kp, ok := ks.rsa[name]
	if !ok {
		return nil, &KeyNotFoundError{Name: name, Type: "RSA"}
	}
	return kp, nil
}

// ECDSA 获取指定名称的 ECDSA 密钥对
func (ks *KeyStore) ECDSA(name string) (*ECDSAKeyPair, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	kp, ok := ks.ecdsa[name]
	if !ok {
		return nil, &KeyNotFoundError{Name: name, Type: "ECDSA"}
	}
	return kp, nil
}

// HasRSA 检查是否存在指定名称的 RSA 密钥
func (ks *KeyStore) HasRSA(name string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	_, ok := ks.rsa[name]
	return ok
}

// HasECDSA 检查是否存在指定名称的 ECDSA 密钥
func (ks *KeyStore) HasECDSA(name string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	_, ok := ks.ecdsa[name]
	return ok
}

// ListRSA 列出所有 RSA 密钥名称
func (ks *KeyStore) ListRSA() []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	names := make([]string, 0, len(ks.rsa))
	for name := range ks.rsa {
		names = append(names, name)
	}
	return names
}

// ListECDSA 列出所有 ECDSA 密钥名称
func (ks *KeyStore) ListECDSA() []string {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	names := make([]string, 0, len(ks.ecdsa))
	for name := range ks.ecdsa {
		names = append(names, name)
	}
	return names
}
