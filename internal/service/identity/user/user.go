package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iVampireSP/go-template/ent"
	entuser "github.com/iVampireSP/go-template/ent/user"
	"github.com/iVampireSP/go-template/internal/infra/cache"
	"github.com/iVampireSP/go-template/internal/infra/jwt"
	"github.com/iVampireSP/go-template/pkg/json"
	"github.com/iVampireSP/go-template/pkg/paginator"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	// MaxFailedLoginAttempts is the maximum number of failed login attempts before lockout.
	MaxFailedLoginAttempts = 5
	// AccountLockDuration is how long an account stays locked after max failed attempts.
	AccountLockDuration = 30 * time.Minute
	// AccessTokenTTL is the default access token lifetime.
	AccessTokenTTL = 2 * time.Hour
)

// User 用户业务服务
type User struct {
	client *ent.Client
	jwt    *jwt.JWT
	locker *cache.Locker
	redis  redis.UniversalClient
}

// NewUser creates a new user service.
func NewUser(
	client *ent.Client,
	jwtSvc *jwt.JWT,
	locker *cache.Locker,
	redisClient redis.UniversalClient,
) *User {
	return &User{
		client: client,
		jwt:    jwtSvc,
		locker: locker,
		redis:  redisClient,
	}
}

// ==================== 认证 ====================

// Authenticate 验证用户凭证
func (s *User) Authenticate(ctx context.Context, email, password string) (*ent.User, error) {
	u, err := s.client.User.Query().
		Where(entuser.EmailEQ(email), entuser.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if u.Status != entuser.StatusActive {
		return nil, ErrUserInactive
	}
	if u.FailedLoginCount >= MaxFailedLoginAttempts && u.FailedLoginAt != nil {
		if time.Since(*u.FailedLoginAt) < AccountLockDuration {
			return nil, ErrUserLocked
		}
		s.resetFailedLogin(ctx, u.ID)
		u.FailedLoginCount = 0
	}
	if u.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)) != nil {
		s.incrementFailedLogin(ctx, u.ID)
		return nil, ErrInvalidPassword
	}

	return u, nil
}

// IssueToken 为用户签发 JWT access token
func (s *User) IssueToken(ctx context.Context, userID int) (string, error) {
	now := time.Now()
	claims := map[string]any{
		"sub":          strconv.Itoa(userID),
		"subject_type": "user",
		"exp":          now.Add(AccessTokenTTL).Unix(),
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
	}
	return s.jwt.GenerateTokenWithCustomClaims(claims)
}

// ResolveToken 验证 JWT token，返回 userID
func (s *User) ResolveToken(ctx context.Context, token string) (int, error) {
	claims, err := s.jwt.ParseMapClaims(token)
	if err != nil {
		return 0, ErrInvalidToken
	}
	sub, _ := claims["sub"].(string)
	userID, err := strconv.Atoi(sub)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}

// RecordLogin 记录登录信息
func (s *User) RecordLogin(ctx context.Context, userID int, clientIP string) error {
	now := time.Now()
	builder := s.client.User.UpdateOneID(userID).
		SetLastLoginAt(now).
		AddLoginCount(1).
		SetFailedLoginCount(0).
		ClearFailedLoginAt()

	if clientIP != "" {
		builder = builder.SetLastLoginIP(clientIP)
	}

	_, err := builder.Save(ctx)
	if err == nil {
		s.invalidateEntityCache(ctx, userID)
	}
	return err
}

// VerifyPassword 验证用户密码（敏感操作确认）
func (s *User) VerifyPassword(ctx context.Context, userID int, password string) error {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)) != nil {
		return ErrInvalidPassword
	}
	return nil
}

// ==================== 用户 CRUD ====================

// GetByID returns a user by ID (excludes soft-deleted).
func (s *User) GetByID(ctx context.Context, userID int) (*ent.User, error) {
	u, err := s.client.User.Query().
		Where(entuser.ID(userID), entuser.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	s.setEntityCache(ctx, u)
	return u, nil
}

// GetByIDCached 带缓存的 GetByID
func (s *User) GetByIDCached(ctx context.Context, userID int) (*ent.User, error) {
	if u, ok := s.getEntityCache(ctx, userID); ok {
		return u, nil
	}
	return s.GetByID(ctx, userID)
}

// GetByEmail returns a user by email (excludes soft-deleted).
func (s *User) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	u, err := s.client.User.Query().
		Where(entuser.EmailEQ(email), entuser.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// Create creates a new user.
func (s *User) Create(ctx context.Context, email, password, displayName, registerIP string) (*ent.User, error) {
	exists, err := s.client.User.Query().
		Where(entuser.EmailEQ(email), entuser.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	creator := s.client.User.Create().
		SetEmail(email).
		SetPasswordHash(string(hash)).
		SetDisplayName(displayName).
		SetStatus(entuser.StatusActive)
	if registerIP != "" {
		creator.SetRegisterIP(registerIP)
	}
	u, err := creator.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return u, nil
}

// UpdateProfileInput 更新用户资料输入
type UpdateProfileInput struct {
	DisplayName *string
	AvatarURL   *string
}

// UpdateProfile 更新用户资料
func (s *User) UpdateProfile(ctx context.Context, userID int, input UpdateProfileInput) (*ent.User, error) {
	if input.DisplayName == nil && input.AvatarURL == nil {
		return s.GetByID(ctx, userID)
	}

	updater := s.client.User.UpdateOneID(userID)

	if input.DisplayName != nil {
		if *input.DisplayName == "" {
			return nil, ErrInvalidDisplayName
		}
		updater = updater.SetDisplayName(*input.DisplayName)
	}

	if input.AvatarURL != nil {
		if *input.AvatarURL == "" {
			updater = updater.ClearAvatarURL()
		} else {
			updater = updater.SetAvatarURL(*input.AvatarURL)
		}
	}

	u, err := updater.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	s.invalidateEntityCache(ctx, userID)
	return u, nil
}

// UpdatePassword updates a user's password.
func (s *User) UpdatePassword(ctx context.Context, userID int, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	_, err = s.client.User.UpdateOneID(userID).
		SetPasswordHash(string(hash)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrUserNotFound
		}
		return err
	}
	s.invalidateEntityCache(ctx, userID)
	return nil
}

// UpdateStatus updates a user's status.
func (s *User) UpdateStatus(ctx context.Context, userID int, status string) error {
	err := s.client.User.UpdateOneID(userID).
		SetStatus(entuser.Status(status)).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrUserNotFound
		}
		return err
	}
	s.invalidateEntityCache(ctx, userID)
	return nil
}

// Delete 软删除用户
func (s *User) Delete(ctx context.Context, userID int) error {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	now := time.Now()
	deletedEmail := fmt.Sprintf("%s_deleted_%d", u.Email, u.ID)
	if err := s.client.User.UpdateOneID(userID).
		SetDeletedAt(now).
		SetEmail(deletedEmail).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to soft-delete user: %w", err)
	}
	s.invalidateEntityCache(ctx, userID)
	return nil
}

// UserListInput 用户列表筛选条件
type UserListInput struct {
	Search        string
	Status        string
	EmailVerified string
	Page          int
	PerPage       int
}

// List 按条件列出用户
func (s *User) List(ctx context.Context, input UserListInput) ([]*ent.User, int, error) {
	query := s.client.User.Query().Where(entuser.DeletedAtIsNil())
	if input.Search != "" {
		query = query.Where(entuser.Or(
			entuser.EmailContains(input.Search),
			entuser.DisplayNameContains(input.Search),
		))
	}
	if input.Status != "" {
		query = query.Where(entuser.StatusEQ(entuser.Status(input.Status)))
	}
	switch input.EmailVerified {
	case "true":
		query = query.Where(entuser.EmailVerifiedEQ(true))
	case "false":
		query = query.Where(entuser.EmailVerifiedEQ(false))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	users, err := query.
		Order(ent.Desc(entuser.FieldCreatedAt)).
		Offset(paginator.Offset(input.Page, input.PerPage)).
		Limit(input.PerPage).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ==================== 内部方法 ====================

func (s *User) entityCacheKey(userID int) string {
	return fmt.Sprintf("user:entity:%d", userID)
}

func (s *User) setEntityCache(ctx context.Context, u *ent.User) {
	if s.redis == nil || u == nil || u.ID <= 0 {
		return
	}
	body, err := json.Marshal(u)
	if err != nil {
		return
	}
	_ = s.redis.Set(ctx, s.entityCacheKey(u.ID), body, 30*time.Second).Err()
}

func (s *User) getEntityCache(ctx context.Context, userID int) (*ent.User, bool) {
	if s.redis == nil || userID <= 0 {
		return nil, false
	}
	body, err := s.redis.Get(ctx, s.entityCacheKey(userID)).Bytes()
	if err != nil {
		return nil, false
	}
	var u ent.User
	if err := json.Unmarshal(body, &u); err != nil {
		_ = s.redis.Del(ctx, s.entityCacheKey(userID)).Err()
		return nil, false
	}
	return &u, true
}

func (s *User) invalidateEntityCache(ctx context.Context, userID int) {
	if s.redis == nil || userID <= 0 {
		return
	}
	_ = s.redis.Del(ctx, s.entityCacheKey(userID)).Err()
}

func (s *User) incrementFailedLogin(ctx context.Context, userID int) {
	_, _ = s.client.User.UpdateOneID(userID).
		AddFailedLoginCount(1).
		SetFailedLoginAt(time.Now()).
		Save(ctx)
}

func (s *User) resetFailedLogin(ctx context.Context, userID int) {
	_, _ = s.client.User.UpdateOneID(userID).
		SetFailedLoginCount(0).
		ClearFailedLoginAt().
		Save(ctx)
}
