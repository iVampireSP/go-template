package schema

import (
	"time"
)

type UserTokenInfo struct {
	Aud              string           `json:"aud"`
	Iss              string           `json:"iss"`
	Iat              float64          `json:"iat"`
	Exp              float64          `json:"exp"`
	Sub              UserId           `json:"sub" mapstructure:"-"`
	Scopes           []string         `json:"scopes"`
	Roles            *UserRoles       `json:"roles"`
	Permissions      *UserPermissions `json:"permissions"`
	Uuid             string           `json:"uuid"`
	Avatar           string           `json:"avatar"`
	Name             string           `json:"name"`
	EmailVerified    bool             `json:"email_verified"`
	RealNameVerified bool             `json:"real_name_verified"`
	PhoneVerified    bool             `json:"phone_verified"`
	Email            string           `json:"email"`
	Phone            string           `json:"phone"`
	CreatedAt        time.Time        `json:"created_at"`
}

type User struct {
	Token UserTokenInfo
	Valid bool
}

type UserRole string
type UserPermission struct{}

type UserRoles []UserRole
type UserPermissions []UserPermission

type UserId string

func (u *User) GetId() UserId {
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

func (u *User) GetRoles() *UserRoles {
	return u.Token.Roles
}

func (u *User) GetPermissions() *UserPermissions {
	return u.Token.Permissions
}

func (u UserId) String() string {
	return string(u)
}

func (ur *UserRoles) Has(role UserRole) bool {
	for _, r := range *ur {
		if r == role {
			return true
		}
	}

	return false
}

func (up *UserPermissions) Has(permission UserPermission) bool {
	for _, p := range *up {
		if p == permission {
			return true
		}
	}

	return false
}
