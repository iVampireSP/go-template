package constants

type JwtTokenTypes string

const (
	JwtTokenTypeAccessToken JwtTokenTypes = "access_token"
	JwtTokenTypeIDToken     JwtTokenTypes = "id_token"
)

func (jwtTokenType JwtTokenTypes) String() string {
	return string(jwtTokenType)
}

const (
	AuthHeader = "Authorization"
	AuthPrefix = "Bearer"

	AuthMiddlewareKey               = "auth.user"
	AuthAssistantShareMiddlewareKey = "auth.assistant.share"
)
