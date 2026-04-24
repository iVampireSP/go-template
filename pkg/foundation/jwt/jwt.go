package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/keystore"
)

const (
	// RefreshWindowDuration token 刷新窗口（只有在到期前这段时间内才能刷新）
	RefreshWindowDuration = 30 * time.Minute
)

// JWT JWT服务
type JWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiresIn  int // 秒
}

// Claims JWT声明
type Claims struct {
	UserID      int    `json:"user_id"`
	SubjectType string `json:"subject_type"`
	jwt.RegisteredClaims
}

// NewJWT 从配置创建 JWT 服务
func NewJWT(ks *keystore.KeyStore) (*JWT, error) {
	keyName := config.String("jwt.key", "rsa")

	keyPair, err := ks.RSA(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT key '%s': %w", keyName, err)
	}

	return &JWT{
		privateKey: keyPair.PrivateKey,
		publicKey:  keyPair.PublicKey,
		issuer:     config.String("discovery.issuer", "http://localhost"),
		expiresIn:  config.Int("jwt.expires_in", 86400),
	}, nil
}

// GenerateToken 生成JWT令牌（基础）
func (s *JWT) GenerateToken(userID int, subjectType string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.expiresIn) * time.Second)

	claims := &Claims{
		UserID:      userID,
		SubjectType: subjectType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // JTI
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{s.issuer},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.getKeyID()

	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenWithExtraClaims 生成带额外业务字段的 JWT（业务层自行决定字段含义）
func (s *JWT) GenerateTokenWithExtraClaims(userID int, subjectType string, extra map[string]any) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.expiresIn) * time.Second)

	claims := jwt.MapClaims{
		"jti":          uuid.New().String(), // JTI
		"user_id":      userID,
		"subject_type": subjectType,
		"iss":          s.issuer,
		"sub":          fmt.Sprintf("%d", userID),
		"aud":          []string{s.issuer},
		"exp":          expiresAt.Unix(),
		"nbf":          now.Unix(),
		"iat":          now.Unix(),
	}
	for k, v := range extra {
		// 避免覆盖基础字段
		if k == "jti" || k == "user_id" || k == "subject_type" || k == "iss" || k == "sub" || k == "aud" || k == "exp" || k == "nbf" || k == "iat" {
			continue
		}
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.getKeyID()

	str, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return str, nil
}

// ValidateToken 验证JWT令牌
func (s *JWT) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// 解析为 MapClaims 再提取基础字段
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	var out Claims
	// JTI
	if v, ok := mc["jti"]; ok {
		if str, ok2 := v.(string); ok2 {
			out.ID = str
		}
	}
	// UserID
	if v, ok := mc["user_id"]; ok {
		switch vv := v.(type) {
		case float64:
			out.UserID = int(vv)
		case int:
			out.UserID = int(vv)
		case int64:
			out.UserID = int(vv)
		}
	}
	// SubjectType
	if v, ok := mc["subject_type"]; ok {
		if str, ok2 := v.(string); ok2 {
			out.SubjectType = str
		}
	}
	// 提取标准时间字段
	if v, ok := mc["exp"]; ok {
		if f, ok2 := v.(float64); ok2 {
			out.ExpiresAt = jwt.NewNumericDate(time.Unix(int64(f), 0))
		}
	}
	if v, ok := mc["iat"]; ok {
		if f, ok2 := v.(float64); ok2 {
			out.IssuedAt = jwt.NewNumericDate(time.Unix(int64(f), 0))
		}
	}
	if v, ok := mc["nbf"]; ok {
		if f, ok2 := v.(float64); ok2 {
			out.NotBefore = jwt.NewNumericDate(time.Unix(int64(f), 0))
		}
	}
	if v, ok := mc["iss"]; ok {
		if str, ok2 := v.(string); ok2 {
			out.Issuer = str
		}
	}
	return &out, nil
}

// ParseMapClaims validates token and returns full map claims.
func (s *JWT) ParseMapClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// ParseToken 解析 token 并返回 userID 和 subjectType
func (s *JWT) ParseToken(tokenString string) (int, string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, "", err
	}

	return claims.UserID, claims.SubjectType, nil
}

// RefreshToken 刷新JWT令牌（保留原有所有自定义字段，生成新 JTI）
func (s *JWT) RefreshToken(tokenString string) (string, error) {
	// 解析原 token 为 MapClaims
	orig, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil || !orig.Valid {
		return "", fmt.Errorf("failed to parse token for refresh: %w", err)
	}
	mc, ok := orig.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// 检查刷新窗口（30分钟内到期才可刷新）
	now := time.Now()
	var expUnix int64
	if v, ok := mc["exp"].(float64); ok {
		expUnix = int64(v)
	}
	expiresAt := time.Unix(expUnix, 0)
	if expiresAt.Sub(now) > RefreshWindowDuration {
		return "", fmt.Errorf("token is not eligible for refresh yet")
	}

	// 更新时间字段和 JTI，重新签名
	newExp := now.Add(time.Duration(s.expiresIn) * time.Second).Unix()
	mc["jti"] = uuid.New().String() // 新 JTI
	mc["exp"] = newExp
	mc["iat"] = now.Unix()
	mc["nbf"] = now.Unix()

	newToken := jwt.NewWithClaims(jwt.SigningMethodRS256, mc)
	newToken.Header["kid"] = s.getKeyID()

	str, err := newToken.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refreshed token: %w", err)
	}
	return str, nil
}

// GetPublicKey 获取公钥（用于其他服务验证JWT）
func (s *JWT) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}

// getKeyID 生成密钥ID（与 JWKS handler 使用完全相同的算法）
func (s *JWT) getKeyID() string {
	// 必须与 jwks_handler.go 的 generateKeyID 完全一致
	hash := sha256.New()
	hash.Write(s.publicKey.N.Bytes())
	hash.Write(big.NewInt(int64(s.publicKey.E)).Bytes())

	hashBytes := hash.Sum(nil)
	// 使用 base64 URL 编码（与 JWKS 一致）
	return base64.RawURLEncoding.EncodeToString(hashBytes[:16])
}

// GenerateTokenWithCustomClaims 生成完全自定义声明的 JWT（用于 OIDC ID Token）
func (s *JWT) GenerateTokenWithCustomClaims(claims map[string]any) (string, error) {
	mapClaims := jwt.MapClaims{}
	for k, v := range claims {
		mapClaims[k] = v
	}
	// 确保有 JTI
	if _, ok := mapClaims["jti"]; !ok {
		mapClaims["jti"] = uuid.New().String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)
	token.Header["kid"] = s.getKeyID()

	str, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return str, nil
}

// GetIssuer 获取发行者
func (s *JWT) GetIssuer() string {
	return s.issuer
}

// GetExpiresIn 获取令牌过期时间（秒）
func (s *JWT) GetExpiresIn() int {
	return s.expiresIn
}
