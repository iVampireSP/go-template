package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iVampireSP/go-template/internal/api/admin/v1/resource"
	"github.com/iVampireSP/go-template/internal/api/admin/v1/response"
	"github.com/iVampireSP/go-template/internal/api/admin/v1/request"
	"github.com/iVampireSP/go-template/internal/service/identity/admin"
	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/redis/go-redis/v9"
)

// AuthHandler 管理员认证handler
type AuthHandler struct {
	adminService *admin.Admin
	redisClient  redis.UniversalClient
}

// NewAuthHandler 创建管理员认证handler
func NewAuthHandler(adminService *admin.Admin, redisClient redis.UniversalClient) *AuthHandler {
	return &AuthHandler{
		adminService: adminService,
		redisClient:  redisClient,
	}
}

// Login 管理员登录
func (c *AuthHandler) Login(ctx context.Context, input *request.LoginRequest) (*response.LoginResponse, error) {
	clientIP := httpserver.GetClientIP(ctx).String()

	if err := c.checkLoginRateLimit(ctx, input.Body.Email, clientIP); err != nil {
		return nil, err
	}

	a, err := c.adminService.Authenticate(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		c.recordLoginFailure(ctx, input.Body.Email, clientIP)
		if errors.Is(err, admin.ErrAdminNotFound) || errors.Is(err, admin.ErrInvalidPassword) {
			return nil, cerr.Unauthorized("invalid credentials")
		}
		return nil, err
	}
	c.clearLoginAccountLimit(ctx, input.Body.Email)

	token, err := c.adminService.IssueToken(ctx, a.ID)
	if err != nil {
		return nil, err
	}

	_ = c.adminService.RecordLogin(ctx, a.ID, clientIP)

	loginResource := resource.LoginResource{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(admin.AccessTokenTTL.Seconds()),
	}
	if a != nil {
		loginResource.Admin = &resource.ProfileResource{
			ID:          a.ID,
			Email:       a.Email,
			DisplayName: a.DisplayName,
			Status:      string(a.Status),
			CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return response.NewLoginResponse(loginResource), nil
}

// Me 获取当前管理员信息
func (c *AuthHandler) Me(ctx context.Context, _ *request.MeRequest) (*response.ProfileResponse, error) {
	a := admin.MustContext(ctx)
	return response.NewProfileResponse(resource.ProfileResource{
		ID:          a.ID,
		Email:       a.Email,
		DisplayName: a.DisplayName,
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
	}), nil
}

// ==================== Rate Limiting ====================

func (c *AuthHandler) checkLoginRateLimit(ctx context.Context, email, clientIP string) error {
	const (
		window       = time.Minute
		ipLimit      = 10
		accountLimit = 5
	)
	if exceeded(ctx, c.redisClient, fmt.Sprintf("auth:login:admin:ip:%s", clientIP), ipLimit, window) {
		return cerr.TooManyRequests("too many login attempts from this ip")
	}
	if exceeded(ctx, c.redisClient, fmt.Sprintf("auth:login:admin:account:%s", strings.ToLower(email)), accountLimit, window) {
		return cerr.TooManyRequests("too many login attempts for this account")
	}
	return nil
}

func (c *AuthHandler) recordLoginFailure(ctx context.Context, email, clientIP string) {
	const window = time.Minute
	incrementWithTTL(ctx, c.redisClient, fmt.Sprintf("auth:login:admin:ip:%s", clientIP), window)
	incrementWithTTL(ctx, c.redisClient, fmt.Sprintf("auth:login:admin:account:%s", strings.ToLower(email)), window)
}

func (c *AuthHandler) clearLoginAccountLimit(ctx context.Context, email string) {
	_ = c.redisClient.Del(ctx, fmt.Sprintf("auth:login:admin:account:%s", strings.ToLower(email))).Err()
}

func incrementWithTTL(ctx context.Context, client redis.UniversalClient, key string, ttl time.Duration) {
	if client == nil {
		return
	}
	val, err := client.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if val == 1 {
		_ = client.Expire(ctx, key, ttl).Err()
	}
}

func exceeded(ctx context.Context, client redis.UniversalClient, key string, limit int64, ttl time.Duration) bool {
	if client == nil {
		return false
	}
	val, _ := client.Get(ctx, key).Int64()
	if val < limit {
		return false
	}
	_ = client.Expire(ctx, key, ttl).Err()
	return true
}
