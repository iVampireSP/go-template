package schema

type JWTTokenTypes string

const (
	JWTAccessToken JWTTokenTypes = "access_token"
	JWTIDToken     JWTTokenTypes = "id_token"
)

func (jwtTokenTypes JWTTokenTypes) String() string {
	return string(jwtTokenTypes)
}
