package jwt

// Config holds JWT signing and validation parameters.
type Config struct {
	KeyName   string // keystore key name, default "rsa"
	Issuer    string // token issuer URL
	ExpiresIn int    // token lifetime in seconds
}
