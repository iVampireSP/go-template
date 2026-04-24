package handler

import (
	"net/http"
	"strings"

	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/json"
)

type DiscoveryHandler struct {
	jwks *JWKSHandler
}

func NewDiscoveryHandler(jwks *JWKSHandler) *DiscoveryHandler {
	return &DiscoveryHandler{
		jwks: jwks,
	}
}

type discoveryResponse struct {
	Issuer          string `json:"issuer"`
	JWKSURI         string `json:"jwks_uri"`
	UserAPIBaseURL  string `json:"user_api_base_url,omitempty"`
	AdminAPIBaseURL string `json:"admin_api_base_url,omitempty"`
}

func (h *DiscoveryHandler) GetUserDiscovery(w http.ResponseWriter, _ *http.Request) {
	baseURL := config.String("discovery.user_api_base_url")
	resp := discoveryResponse{
		Issuer:         realmIssuer("user"),
		JWKSURI:        joinURL(baseURL, "/.well-known/jwks.json"),
		UserAPIBaseURL: baseURL,
	}
	writeJSON(w, resp)
}

func (h *DiscoveryHandler) GetAdminDiscovery(w http.ResponseWriter, _ *http.Request) {
	baseURL := config.String("discovery.admin_api_base_url")
	resp := discoveryResponse{
		Issuer:          realmIssuer("admin"),
		JWKSURI:         joinURL(baseURL, "/.well-known/jwks.json"),
		AdminAPIBaseURL: baseURL,
	}
	writeJSON(w, resp)
}

func (h *DiscoveryHandler) GetUserOpenIDConfiguration(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "not configured"})
}

func (h *DiscoveryHandler) GetAdminOpenIDConfiguration(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "not configured"})
}

func writeJSON(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}

func joinURL(base, suffix string) string {
	if base == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + suffix
}

func realmIssuer(subjectType string) string {
	base := config.String("discovery.issuer")
	if subjectType == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/realms/" + subjectType
}
