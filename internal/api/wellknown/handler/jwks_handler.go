package handler

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"net/http"
	"sync"

	"github.com/iVampireSP/go-template/pkg/json"

	"github.com/iVampireSP/go-template/internal/infra/jwt"
)

// JWKSHandler handles JWKS (JSON Web Key Set) request.
type JWKSHandler struct {
	jwtService *jwt.JWT
	cachedJWKS *JWKS
	cacheMutex sync.RWMutex
}

// NewJWKSHandler creates a new JWKS handler.
func NewJWKSHandler(jwtService *jwt.JWT) *JWKSHandler {
	handler := &JWKSHandler{jwtService: jwtService}
	handler.initCache()
	return handler
}

func (h *JWKSHandler) initCache() {
	publicKey := h.jwtService.GetPublicKey()
	if publicKey == nil {
		return
	}

	jwk := rsaPublicKeyToJWK(publicKey)
	h.cacheMutex.Lock()
	h.cachedJWKS = &JWKS{Keys: []JWK{jwk}}
	h.cacheMutex.Unlock()
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// GetJWKS returns the platform's JWKS as http.HandlerFunc.
func (h *JWKSHandler) GetJWKS(w http.ResponseWriter, _ *http.Request) {
	h.cacheMutex.RLock()
	cachedJWKS := h.cachedJWKS
	h.cacheMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if cachedJWKS == nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "JWKS not available"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cachedJWKS)
}

func rsaPublicKeyToJWK(publicKey *rsa.PublicKey) JWK {
	kid := generateKeyID(publicKey)
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}

// SigningAlgorithms returns the signing algorithms from the cached JWKS keys.
func (h *JWKSHandler) SigningAlgorithms() []string {
	h.cacheMutex.RLock()
	defer h.cacheMutex.RUnlock()
	if h.cachedJWKS == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var algs []string
	for _, k := range h.cachedJWKS.Keys {
		if k.Alg != "" {
			if _, ok := seen[k.Alg]; !ok {
				seen[k.Alg] = struct{}{}
				algs = append(algs, k.Alg)
			}
		}
	}
	return algs
}

func generateKeyID(publicKey *rsa.PublicKey) string {
	hash := sha256.New()
	_, _ = hash.Write(publicKey.N.Bytes())
	_, _ = hash.Write(big.NewInt(int64(publicKey.E)).Bytes())
	hashBytes := hash.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hashBytes[:16])
}
