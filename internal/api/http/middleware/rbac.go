package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go-template/internal/api/http/response"
	"go-template/internal/base/conf"
	userPkg "go-template/internal/pkg/user"
	"go-template/internal/service/auth"
	"net/http"
	"strings"
)

type RBAC struct {
	authService *auth.Service
	config      *conf.Config
}

func (m *RBAC) RoutePermission() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := m.authService.GetUserSafe(c)
		if !ok {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		if !user.Valid {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		var path = cleanPath(c.Path())
		act := strings.ToLower(c.Method())

		permissionName := userPkg.Permission(m.config.App.Name + ":" + path + ":" + act)
		pass := user.HasPermissions(permissionName)

		if !pass {
			return response.Ctx(c).
				Message(fmt.Sprintf("permission denied, permission name: %s", permissionName)).
				Error(nil).
				Status(http.StatusForbidden).
				Send()
		}

		return c.Next()
	}
}

func (m *RBAC) RequirePermissions(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := m.authService.GetUserSafe(c)
		if !ok {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		if !user.Valid {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		var pass = true
		var failedPermissionName string

		for _, permission := range permissions {
			permissionName := userPkg.Permission(m.config.App.Name + ":" + permission)

			has := user.HasPermissions(permissionName)
			if !has {
				failedPermissionName = permissionName.String()
				pass = false
				break
			}
		}

		if !pass {
			return response.Ctx(c).
				Message(fmt.Sprintf("permission denied, required permissions: %s, failed permission: %s",
					permissions, failedPermissionName)).
				Error(nil).
				Status(http.StatusForbidden).
				Send()
		}

		return c.Next()
	}
}

func (m *RBAC) RequireRoles(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := m.authService.GetUserSafe(c)
		if !ok {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		if !user.Valid {
			return response.Ctx(c).Error(nil).Status(http.StatusUnauthorized).Send()
		}

		var pass = true
		var failedRoleName string

		for _, role := range roles {
			roleName := userPkg.Role(m.config.App.Name + ":" + role)
			pass = user.HasRoles(roleName)

			if !pass {
				failedRoleName = roleName.String()
				break
			}

		}

		if !pass {
			return response.Ctx(c).
				Message(fmt.Sprintf("permission denied, required roles: %s, failed role %s", roles, failedRoleName)).
				Error(nil).
				Status(http.StatusForbidden).
				Send()
		}

		return c.Next()
	}
}

func NewRBAC(authService *auth.Service, config *conf.Config) *RBAC {
	return &RBAC{
		authService: authService,
		config:      config,
	}
}

func cleanPath(path string) string {
	// 如果第一个字符是 /，则删掉
	if path[0] == '/' {
		path = path[1:]
	}

	// 将所有的 / 转为 :
	return strings.ReplaceAll(path, "/", ":")

}
