package user

import (
	"slices"
	"time"
)

type Token struct {
	Aud              string       `json:"aud"`
	Iss              string       `json:"iss"`
	Iat              float64      `json:"iat"`
	Exp              float64      `json:"exp"`
	Sub              Id           `json:"sub" mapstructure:"-"`
	Scopes           []string     `json:"scopes"`
	Roles            []Role       `json:"roles,omitempty"`
	Permissions      []Permission `json:"permissions"`
	Uuid             string       `json:"uuid"`
	Avatar           string       `json:"avatar"`
	Name             string       `json:"name"`
	EmailVerified    bool         `json:"email_verified"`
	RealNameVerified bool         `json:"real_name_verified"`
	PhoneVerified    bool         `json:"phone_verified"`
	Email            string       `json:"email"`
	Phone            string       `json:"phone"`
	CreatedAt        time.Time    `json:"created_at"`
}

type User struct {
	Token Token
	Id    Id
	Valid bool
}

type Role string

func (r Role) String() string {
	return string(r)
}

type Permission string

type Id string

func (u *User) GetId() Id {
	return u.Token.Sub
}

func (u *User) GetName() string {
	return u.Token.Name
}

func (u *User) GetEmail() string {
	return u.Token.Email
}

func (u *User) GetPhone() string {
	return u.Token.Phone
}

func (u *User) GetRoles() []Role {
	return u.Token.Roles
}

func (u *User) GetPermissions() []Permission {
	return u.Token.Permissions
}

func (u *User) GetAvatar() string {
	return u.Token.Avatar
}

func (u *User) GetUuid() string {
	return u.Token.Uuid
}

func (u *User) HasRoles(roles ...Role) bool {
	if len(roles) == 0 {
		return true
	}

	for _, role := range roles {
		pass := slices.Contains(u.Token.Roles, role)

		if !pass {
			return false
		}
	}

	return true
}

func (u *User) HasPermissions(permissions ...Permission) bool {
	if len(permissions) == 0 {
		return true
	}

	for _, p := range permissions {
		pass := slices.Contains(u.Token.Permissions, p)

		if !pass {
			return false
		}
	}

	return true
}

func (u Id) String() string {
	return string(u)
}

func (up Permission) String() string {
	return string(up)
}
