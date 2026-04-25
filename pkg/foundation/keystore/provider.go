package keystore

import (
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/container"
)

type Provider struct {
	app *container.Application
}

func NewProvider(app *container.Application) *Provider {
	return &Provider{app: app}
}

func (p *Provider) Register() {
	p.app.Singleton(NewDefaultConfig)
	p.app.Singleton(NewKeyStore)
}

func (p *Provider) Boot() {}

// NewDefaultConfig returns a KeyStore Config populated from the application config.
func NewDefaultConfig() Config {
	keysRaw := config.Map("keystore.keys")
	keys := make(map[string]KeyConfig, len(keysRaw))
	for name, v := range keysRaw {
		kc, ok := v.(map[string]any)
		if !ok {
			continue
		}
		keys[name] = KeyConfig{
			Type:       stringVal(kc, "type"),
			PrivateKey: stringVal(kc, "private_key"),
			PublicKey:  stringVal(kc, "public_key"),
		}
	}
	return Config{Keys: keys}
}

func stringVal(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
