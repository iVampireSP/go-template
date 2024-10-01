package jwks

import (
	"errors"
	"go-template/internal/base/conf"
	"go-template/internal/base/logger"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

var Jwks keyfunc.Keyfunc

var (
	ErrJWKSNotInitialized = errors.New("JWKS is not initialized")
)

type JWKS struct {
	url    string
	logger *logger.Logger
	config *conf.Config
}

func NewJWKS(config *conf.Config, logger *logger.Logger) *JWKS {
	return &JWKS{
		url:    config.JWKS.Url,
		logger: logger,
		config: config,
	}
}

func (j *JWKS) RefreshJWKS() {
	if j.config.Debug.Enabled {
		return
	}

	j.logger.Logger.Info("Refreshing JWKS...")

	var err error

	Jwks, err = keyfunc.NewDefault([]string{j.url})
	if err != nil {
		j.logger.Logger.Error("Failed to create JWK Set from resource at the given URL.\nError: " + err.Error())
	}

	j.logger.Logger.Info("JWKS refreshed.")
}

func (j *JWKS) ParseJWT(jwtB64 string) (*jwt.Token, error) {
	if Jwks.Keyfunc == nil {
		return nil, ErrJWKSNotInitialized
	}

	token, err := jwt.Parse(jwtB64, Jwks.Keyfunc)

	return token, err
}
