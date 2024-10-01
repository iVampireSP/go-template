package schema

import (
	"time"
)

type UserTokenInfo struct {
	Aud              string    `json:"aud"`
	Iss              string    `json:"iss"`
	Iat              float64   `json:"iat"`
	Exp              float64   `json:"exp"`
	Sub              UserId    `json:"sub" mapstructure:"-"`
	Scopes           []string  `json:"scopes"`
	Id               int       `json:"id"`
	Uuid             string    `json:"uuid"`
	Avatar           string    `json:"avatar"`
	Name             string    `json:"name"`
	EmailVerified    bool      `json:"email_verified"`
	RealNameVerified bool      `json:"real_name_verified"`
	PhoneVerified    bool      `json:"phone_verified"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	CreatedAt        time.Time `json:"created_at"`
}

type User struct {
	Token UserTokenInfo
	Valid bool
}

type UserId string

func (u UserId) String() string {
	return string(u)
}

type JWTTokenTypes string

const (
	JWTAccessToken JWTTokenTypes = "access_token"
	JWTIDToken     JWTTokenTypes = "id_token"
)

func (jwtTokenTypes JWTTokenTypes) String() string {
	return string(jwtTokenTypes)
}
