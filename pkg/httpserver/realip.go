package httpserver

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/iVampireSP/go-template/internal/infra/config"
)

// clientIPKey is the context key for storing the resolved client IP.
type clientIPKey struct{}

// RealIPConfig configures the RealIP middleware.
type RealIPConfig struct {
	// TrustedProxies is a list of CIDR ranges to trust for X-Forwarded-For parsing.
	// If the request comes from a trusted proxy, the middleware will parse
	// X-Forwarded-For to find the real client IP.
	TrustedProxies []string `yaml:"trusted_proxies"`

	// AllowedHeaders is the ordered list of trusted headers for client IP.
	AllowedHeaders []string `yaml:"headers"`
}

// DefaultRealIPConfig returns configuration loaded from http.real_ip config.
func DefaultRealIPConfig() RealIPConfig {
	cfg := RealIPConfig{
		AllowedHeaders: config.Strings("trustedproxy.headers"),
	}

	cfg.TrustedProxies = config.Strings("trustedproxy.proxies")

	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{
			"Forwarded",
			"X-Forwarded-For",
			"X-Forwarded-Host",
			"X-Forwarded-Port",
			"X-Forwarded-Proto",
			"X-Forwarded-Prefix",
			"X-Real-IP",
		}
	}

	// Fallback to default if no proxies configured
	if len(cfg.TrustedProxies) == 0 {
		cfg.TrustedProxies = []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"::1/128",
			"169.254.0.0/16",
			"fe80::/10",
		}
	}

	return cfg
}

// RealIP returns a middleware that resolves the real client IP address.
//
// If the request comes from a trusted proxy, priority order is defined
// by the headers list from configuration (default includes CF + X-Forwarded-For).
//
// The resolved IP is stored in the request context and can be retrieved
// using GetClientIP(ctx).
// nonIPHeaders 不携带客户端 IP，在初始化时过滤掉以避免每次请求白跑 parseIP。
var nonIPHeaders = map[string]struct{}{
	"X-Forwarded-Host":   {},
	"X-Forwarded-Port":   {},
	"X-Forwarded-Proto":  {},
	"X-Forwarded-Prefix": {},
}

func RealIP(cfg RealIPConfig) func(http.Handler) http.Handler {
	// Parse trusted proxy CIDRs
	var trustedNets []*net.IPNet
	allowAll := false
	for _, cidr := range cfg.TrustedProxies {
		if strings.TrimSpace(cidr) == "*" {
			allowAll = true
			break
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			trustedNets = append(trustedNets, ipNet)
		}
	}

	// 预计算 canonical header 名并过滤非 IP header，避免每次请求重复计算。
	var ipHeaders []string
	for _, h := range cfg.AllowedHeaders {
		canonical := http.CanonicalHeaderKey(h)
		if _, skip := nonIPHeaders[canonical]; !skip {
			ipHeaders = append(ipHeaders, canonical)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := resolveClientIP(r, ipHeaders, trustedNets, allowAll)

			ctx := context.WithValue(r.Context(), clientIPKey{}, clientIP)
			r = r.WithContext(ctx)

			if clientIP != nil {
				r.RemoteAddr = clientIP.String()
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP retrieves the resolved client IP from the request context.
func GetClientIP(ctx context.Context) net.IP {
	if ip, ok := ctx.Value(clientIPKey{}).(net.IP); ok {
		return ip
	}
	return nil
}

// resolveClientIP determines the real client IP from the request.
func resolveClientIP(r *http.Request, ipHeaders []string, trustedNets []*net.IPNet, allowAll bool) net.IP {
	remoteIP := parseIP(r.RemoteAddr)
	if remoteIP == nil {
		return nil
	}

	if !allowAll && !isTrustedProxy(remoteIP, trustedNets) {
		return remoteIP
	}

	// Iterate headers in config order so the user controls priority.
	for _, h := range ipHeaders {
		switch h {
		case "Forwarded":
			if fwd := r.Header.Get("Forwarded"); fwd != "" {
				if ip := parseForwardedFor(fwd); ip != nil {
					return ip
				}
			}

		case "X-Forwarded-For":
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ips := strings.Split(xff, ",")
				for i := len(ips) - 1; i >= 0; i-- {
					ip := parseIP(strings.TrimSpace(ips[i]))
					if ip == nil {
						continue
					}
					if !isTrustedProxy(ip, trustedNets) {
						return ip
					}
				}
				if ip := parseIP(strings.TrimSpace(ips[0])); ip != nil {
					return ip
				}
			}

		default:
			if val := r.Header.Get(h); val != "" {
				if ip := parseIP(val); ip != nil {
					return ip
				}
			}
		}
	}

	return remoteIP
}

// parseIP extracts and parses an IP from an address string.
// Handles both "ip" and "ip:port" formats.
func parseIP(addr string) net.IP {
	// Try to split host:port
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// No port, use as-is
		host = addr
	}
	return net.ParseIP(strings.TrimSpace(host))
}

// parseForwardedFor parses RFC 7239 Forwarded header and extracts the first "for" value.
func parseForwardedFor(header string) net.IP {
	for part := range strings.SplitSeq(header, ",") {
		for item := range strings.SplitSeq(part, ";") {
			kv := strings.SplitN(strings.TrimSpace(item), "=", 2)
			if len(kv) != 2 {
				continue
			}
			if strings.EqualFold(kv[0], "for") {
				value := strings.Trim(kv[1], "\"")
				if ip := parseIP(value); ip != nil {
					return ip
				}
			}
		}
	}
	return nil
}

// isTrustedProxy checks if the IP is in any of the trusted networks.
func isTrustedProxy(ip net.IP, trustedNets []*net.IPNet) bool {
	for _, n := range trustedNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
