package keystore

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// RSAKeyPair RSA 密钥对
type RSAKeyPair struct {
	PrivateKey    *rsa.PrivateKey
	PublicKey     *rsa.PublicKey
	PrivateKeyPEM []byte
	PublicKeyPEM  []byte
}

// ECDSAKeyPair ECDSA 密钥对 (用于 Web Push VAPID 等)
type ECDSAKeyPair struct {
	PrivateKey    *ecdsa.PrivateKey
	PublicKey     *ecdsa.PublicKey
	PrivateKeyPEM []byte
	PublicKeyPEM  []byte
}

// GenerateRSAKeyPair 生成 RSA 密钥对
func GenerateRSAKeyPair(bits int) (*RSAKeyPair, error) {
	if bits < 2048 {
		bits = 2048
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// 编码私钥为 PEM (PKCS1)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// 编码公钥为 PEM (PKIX)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &RSAKeyPair{
		PrivateKey:    privateKey,
		PublicKey:     &privateKey.PublicKey,
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
	}, nil
}

// GenerateECDSAKeyPair 生成 ECDSA P-256 密钥对 (Web Push VAPID 要求)
func GenerateECDSAKeyPair() (*ECDSAKeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// 编码私钥为 PEM (EC PRIVATE KEY)
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA private key: %w", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码公钥为 PEM (PUBLIC KEY)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &ECDSAKeyPair{
		PrivateKey:    privateKey,
		PublicKey:     &privateKey.PublicKey,
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
	}, nil
}

// ParseRSAKeyPair 从 PEM 解析 RSA 密钥对
func ParseRSAKeyPair(privateKeyPEM, publicKeyPEM []byte) (*RSAKeyPair, error) {
	// 解析私钥
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode private key PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		key, err2 := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse RSA private key (tried PKCS1 and PKCS8): %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA type")
		}
	}

	// 解析公钥
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	if publicKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode public key PEM block")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA type")
	}

	// 验证密钥对匹配
	if privateKey.PublicKey.N.Cmp(publicKey.N) != 0 || privateKey.PublicKey.E != publicKey.E {
		return nil, fmt.Errorf("private key and public key do not match")
	}

	return &RSAKeyPair{
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
	}, nil
}

// ParseECDSAKeyPair 从 PEM 解析 ECDSA 密钥对
func ParseECDSAKeyPair(privateKeyPEM, publicKeyPEM []byte) (*ECDSAKeyPair, error) {
	// 解析私钥
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode private key PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		key, err2 := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not ECDSA type")
		}
	}

	// 解析公钥
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	if publicKeyBlock == nil {
		return nil, fmt.Errorf("failed to decode public key PEM block")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA type")
	}

	// 验证密钥对匹配
	privatePublicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private public key: %w", err)
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parsed public key: %w", err)
	}
	if !bytes.Equal(privatePublicKeyBytes, publicKeyBytes) {
		return nil, fmt.Errorf("private key and public key do not match")
	}

	return &ECDSAKeyPair{
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		PrivateKeyPEM: privateKeyPEM,
		PublicKeyPEM:  publicKeyPEM,
	}, nil
}

// Base64PrivateKey 返回 base64 编码的私钥
func (kp *RSAKeyPair) Base64PrivateKey() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKeyPEM)
}

// Base64PublicKey 返回 base64 编码的公钥
func (kp *RSAKeyPair) Base64PublicKey() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKeyPEM)
}

// Base64PrivateKey 返回 base64 编码的私钥
func (kp *ECDSAKeyPair) Base64PrivateKey() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKeyPEM)
}

// Base64PublicKey 返回 base64 编码的公钥
func (kp *ECDSAKeyPair) Base64PublicKey() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKeyPEM)
}
