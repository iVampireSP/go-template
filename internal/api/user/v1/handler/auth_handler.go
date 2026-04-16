package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iVampireSP/go-template/internal/api/user/v1/request"
	"github.com/iVampireSP/go-template/internal/api/user/v1/resource"
	"github.com/iVampireSP/go-template/internal/api/user/v1/response"
	"github.com/iVampireSP/go-template/internal/service/identity/user"
	"github.com/iVampireSP/go-template/pkg/cerr"
	"github.com/iVampireSP/go-template/pkg/httpserver"
	"github.com/redis/go-redis/v9"
)

// AuthHandler 用户认证handler
type AuthHandler struct {
	userService *user.User
	redisClient redis.UniversalClient
}

// NewAuthHandler 创建用户认证handler
func NewAuthHandler(userService *user.User, redisClient redis.UniversalClient) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		redisClient: redisClient,
	}
}

// Login 用户登录
func (c *AuthHandler) Login(ctx context.Context, input *request.LoginRequest) (*response.TokenResponse, error) {
	clientIP := httpserver.GetClientIP(ctx).String()

	if err := c.checkLoginRateLimit(ctx, input.Body.Email, clientIP); err != nil {
		return nil, err
	}

	u, err := c.userService.Authenticate(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		c.recordLoginFailure(ctx, input.Body.Email, clientIP)
		if errors.Is(err, user.ErrUserNotFound) || errors.Is(err, user.ErrUserInactive) {
			return nil, user.ErrInvalidPassword
		}
		return nil, err
	}
	c.clearLoginAccountLimit(ctx, input.Body.Email)

	token, err := c.userService.IssueToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	_ = c.userService.RecordLogin(ctx, u.ID, clientIP)

	profile := resource.NewProfileResource(u)
	resp := &response.TokenResponse{}
	resp.Body.User = &profile
	resp.Body.AccessToken = token
	resp.Body.TokenType = "Bearer"
	resp.Body.ExpiresIn = int64(user.AccessTokenTTL.Seconds())
	return resp, nil
}

// Register 用户注册
func (c *AuthHandler) Register(ctx context.Context, input *request.RegisterRequest) (*response.TokenResponse, error) {
	clientIP := httpserver.GetClientIP(ctx).String()

	if input.Body.Password != input.Body.PasswordConfirm {
		return nil, cerr.BadRequest("passwords do not match")
	}

	displayName := input.Body.DisplayName
	if displayName == "" {
		if idx := strings.IndexByte(input.Body.Email, '@'); idx > 0 {
			displayName = input.Body.Email[:idx]
		}
	}

	u, err := c.userService.Create(ctx, input.Body.Email, input.Body.Password, displayName, clientIP)
	if err != nil {
		return nil, err
	}

	token, err := c.userService.IssueToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	profile := resource.NewProfileResource(u)
	resp := &response.TokenResponse{}
	resp.Body.User = &profile
	resp.Body.AccessToken = token
	resp.Body.TokenType = "Bearer"
	resp.Body.ExpiresIn = int64(user.AccessTokenTTL.Seconds())
	return resp, nil
}

// ==================== rate limiting helpers ====================

func (c *AuthHandler) checkLoginRateLimit(ctx context.Context, email, clientIP string) error {
	const (
		window       = time.Minute
		ipLimit      = 10
		accountLimit = 5
	)
	if exceeded(ctx, c.redisClient, loginIPKey("user", clientIP), ipLimit, window) {
		return cerr.TooManyRequests("too many login attempts from this ip, please retry later")
	}
	if exceeded(ctx, c.redisClient, loginAccountKey("user", email), accountLimit, window) {
		return cerr.TooManyRequests("too many login attempts for this account, please retry later")
	}
	return nil
}

func (c *AuthHandler) recordLoginFailure(ctx context.Context, email, clientIP string) {
	const window = time.Minute
	incrementWithTTL(ctx, c.redisClient, loginIPKey("user", clientIP), window)
	incrementWithTTL(ctx, c.redisClient, loginAccountKey("user", email), window)
}

func (c *AuthHandler) clearLoginAccountLimit(ctx context.Context, email string) {
	_ = c.redisClient.Del(ctx, loginAccountKey("user", email)).Err()
}

func loginIPKey(domain, clientIP string) string {
	return fmt.Sprintf("auth:login:%s:ip:%s", domain, strings.TrimSpace(clientIP))
}

func loginAccountKey(domain, email string) string {
	return fmt.Sprintf("auth:login:%s:account:%s", domain, strings.ToLower(strings.TrimSpace(email)))
}

func incrementWithTTL(ctx context.Context, client redis.UniversalClient, key string, ttl time.Duration) {
	if client == nil || key == "" {
		return
	}
	value, err := client.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if value == 1 {
		_ = client.Expire(ctx, key, ttl).Err()
	}
}

func exceeded(ctx context.Context, client redis.UniversalClient, key string, limit int64, ttl time.Duration) bool {
	if client == nil || key == "" || limit <= 0 {
		return false
	}
	value, err := client.Get(ctx, key).Int64()
	if err != nil {
		return false
	}
	if value < limit {
		return false
	}
	_ = client.Expire(ctx, key, ttl).Err()
	return true
}
