package user

import (
	"context"
	"strings"

	"github.com/iVampireSP/go-template/ent"
	"github.com/iVampireSP/go-template/pkg/cerr"
)

type authContextKey struct{}
type authScopesKey struct{}

// WithAuth 将认证信息写入 context（authn 中间件调用）
func WithAuth(ctx context.Context, u *ent.User, scopes []string) context.Context {
	ctx = context.WithValue(ctx, authContextKey{}, u)
	ctx = context.WithValue(ctx, authScopesKey{}, scopes)
	return ctx
}

// Authenticated 返回已认证用户
func Authenticated(ctx context.Context) (*ent.User, bool) {
	raw := ctx.Value(authContextKey{})
	if raw == nil {
		return nil, false
	}
	u, ok := raw.(*ent.User)
	return u, ok && u != nil
}

// AuthScopes 返回当前 token 的 scopes
func AuthScopes(ctx context.Context) []string {
	raw := ctx.Value(authScopesKey{})
	if raw == nil {
		return nil
	}
	scopes, _ := raw.([]string)
	return scopes
}

// HasScopes 检查 scopes 是否包含所有 required
func HasScopes(ctx context.Context, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	scopes := AuthScopes(ctx)
	if len(scopes) == 0 {
		return false
	}
	set := map[string]struct{}{}
	for _, s := range scopes {
		set[strings.TrimSpace(s)] = struct{}{}
	}
	for _, req := range required {
		r := strings.TrimSpace(req)
		if r == "" {
			continue
		}
		if _, ok := set[r]; !ok {
			return false
		}
	}
	return true
}

// RequireScopes 检查 scopes，不满足时返回错误
func RequireScopes(ctx context.Context, required ...string) error {
	if HasScopes(ctx, required...) {
		return nil
	}
	return cerr.Forbidden("insufficient_scope")
}

// FromContext retrieves the authenticated user from context.
func FromContext(ctx context.Context) (*ent.User, bool) {
	return Authenticated(ctx)
}

// GetContext retrieves the authenticated user from context.
func GetContext(ctx context.Context) (*ent.User, bool) {
	return Authenticated(ctx)
}

// MustContext retrieves the authenticated user from context or panics.
func MustContext(ctx context.Context) *ent.User {
	u, ok := Authenticated(ctx)
	if !ok {
		panic(cerr.Unauthorized("authenticated user not found in context"))
	}
	return u
}
