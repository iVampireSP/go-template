package keystore

// Config holds the key definitions loaded from configuration.
type Config struct {
	Keys map[string]KeyConfig
}

// KeyConfig defines a single named key.
type KeyConfig struct {
	Type       string // "rsa" or "ecdsa"
	PrivateKey string // PEM or base64-encoded
	PublicKey  string // PEM or base64-encoded
}
