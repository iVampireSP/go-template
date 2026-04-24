package admin

import (
	"context"
	"strconv"
	"time"

	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/foundation/config"
	"github.com/iVampireSP/go-template/pkg/foundation/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTokenTTL = 2 * time.Hour
)

var (
	ErrAdminNotFound   = cerr.NotFound("admin not found").WithCode("ADMIN_NOT_FOUND")
	ErrInvalidPassword = cerr.Unauthorized("invalid password").WithCode("INVALID_PASSWORD")
	ErrInvalidToken    = cerr.Unauthorized("invalid token").WithCode("INVALID_TOKEN")
)

// AdminEntity is a simple in-memory admin entity (no ent schema needed for template).
type AdminEntity struct {
	ID          int
	Email       string
	DisplayName string
	Status      string
	CreatedAt   time.Time
}

type authContextKey struct{}

// WithAuth writes admin info into context.
func WithAuth(ctx context.Context, a *AdminEntity, scopes []string) context.Context {
	return context.WithValue(ctx, authContextKey{}, a)
}

// Authenticated returns the authenticated admin from context.
func Authenticated(ctx context.Context) (*AdminEntity, bool) {
	raw := ctx.Value(authContextKey{})
	if raw == nil {
		return nil, false
	}
	a, ok := raw.(*AdminEntity)
	return a, ok && a != nil
}

// MustContext retrieves the authenticated admin or panics.
func MustContext(ctx context.Context) *AdminEntity {
	a, ok := Authenticated(ctx)
	if !ok {
		panic(cerr.Unauthorized("admin not found in context"))
	}
	return a
}

// Admin 管理员服务（模板版：基于配置的单管理员）
type Admin struct {
	jwt   *jwt.JWT
	admin *AdminEntity
	hash  string // bcrypt hash of password
}

// NewAdmin creates a new admin service. Reads admin credentials from config.
func NewAdmin(jwtSvc *jwt.JWT) *Admin {
	email := config.String("admin.email", "admin@example.com")
	password := config.String("admin.password", "admin123")
	displayName := config.String("admin.display_name", "Admin")

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return &Admin{
		jwt:  jwtSvc,
		hash: string(hash),
		admin: &AdminEntity{
			ID:          1,
			Email:       email,
			DisplayName: displayName,
			Status:      "active",
			CreatedAt:   time.Now(),
		},
	}
}

// Authenticate validates admin credentials.
func (s *Admin) Authenticate(ctx context.Context, email, password string) (*AdminEntity, error) {
	if email != s.admin.Email {
		return nil, ErrAdminNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(s.hash), []byte(password)) != nil {
		return nil, ErrInvalidPassword
	}
	return s.admin, nil
}

// IssueToken creates a JWT for the admin.
func (s *Admin) IssueToken(ctx context.Context, adminID int) (string, error) {
	now := time.Now()
	claims := map[string]any{
		"sub":          strconv.Itoa(adminID),
		"subject_type": "admin",
		"exp":          now.Add(AccessTokenTTL).Unix(),
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
	}
	return s.jwt.GenerateTokenWithCustomClaims(claims)
}

// ResolveToken validates a JWT and returns admin ID + scopes.
func (s *Admin) ResolveToken(ctx context.Context, token string) (int, []string, error) {
	claims, err := s.jwt.ParseMapClaims(token)
	if err != nil {
		return 0, nil, ErrInvalidToken
	}
	sub, _ := claims["sub"].(string)
	adminID, err := strconv.Atoi(sub)
	if err != nil || adminID <= 0 {
		return 0, nil, ErrInvalidToken
	}
	return adminID, nil, nil
}

// GetByIDCached returns the admin entity by ID.
func (s *Admin) GetByIDCached(ctx context.Context, id int) (*AdminEntity, error) {
	if id == s.admin.ID {
		return s.admin, nil
	}
	return nil, ErrAdminNotFound
}

// RecordLogin records a login event (no-op in template).
func (s *Admin) RecordLogin(ctx context.Context, adminID int, clientIP string) error {
	return nil
}
