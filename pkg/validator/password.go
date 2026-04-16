package validator

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/iVampireSP/go-template/pkg/cerr"
)

var (
	ErrPasswordTooShort       = cerr.BadRequest("password must be at least 8 characters long")
	ErrPasswordTooLong        = cerr.BadRequest("password must not exceed 128 characters")
	ErrPasswordNoUppercase    = cerr.BadRequest("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase    = cerr.BadRequest("password must contain at least one lowercase letter")
	ErrPasswordNoNumber       = cerr.BadRequest("password must contain at least one number")
	ErrPasswordNoSpecialChar  = cerr.BadRequest("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	ErrPasswordCommonPassword = cerr.BadRequest("password is too common, please choose a stronger password")
	ErrPasswordContainsSpaces = cerr.BadRequest("password must not contain spaces")
)

// PasswordStrengthConfig 密码强度配置
type PasswordStrengthConfig struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	NoSpaces       bool
	CheckCommon    bool
}

// DefaultPasswordConfig 默认密码配置
func DefaultPasswordConfig() *PasswordStrengthConfig {
	return &PasswordStrengthConfig{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
		NoSpaces:       true,
		CheckCommon:    true,
	}
}

// WeakPasswordConfig 弱密码配置（用于特殊场景）
func WeakPasswordConfig() *PasswordStrengthConfig {
	return &PasswordStrengthConfig{
		MinLength:      6,
		MaxLength:      128,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false,
		NoSpaces:       true,
		CheckCommon:    false,
	}
}

// 常见弱密码列表（可以从文件加载更多）
var commonPasswords = []string{
	"password", "123456", "12345678", "qwerty", "abc123", "monkey",
	"1234567", "letmein", "trustno1", "dragon", "baseball", "111111",
	"iloveyou", "master", "sunshine", "ashley", "bailey", "passw0rd",
	"shadow", "123123", "654321", "superman", "qazwsx", "michael",
	"football", "admin", "welcome", "login", "princess", "solo",
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string, config *PasswordStrengthConfig) error {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	// 长度检查
	if len(password) < config.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > config.MaxLength {
		return ErrPasswordTooLong
	}

	// 空格检查
	if config.NoSpaces {
		for _, char := range password {
			if unicode.IsSpace(char) {
				return ErrPasswordContainsSpaces
			}
		}
	}

	// 大写字母检查
	if config.RequireUpper {
		hasUpper := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return ErrPasswordNoUppercase
		}
	}

	// 小写字母检查
	if config.RequireLower {
		hasLower := false
		for _, char := range password {
			if unicode.IsLower(char) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return ErrPasswordNoLowercase
		}
	}

	// 数字检查
	if config.RequireNumber {
		hasNumber := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return ErrPasswordNoNumber
		}
	}

	// 特殊字符检查
	if config.RequireSpecial {
		specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`)
		if !specialChars.MatchString(password) {
			return ErrPasswordNoSpecialChar
		}
	}

	// 常见密码检查
	if config.CheckCommon {
		lowerPassword := toLower(password)
		for _, common := range commonPasswords {
			if lowerPassword == common {
				return ErrPasswordCommonPassword
			}
		}
	}

	return nil
}

// toLower 转换为小写（用于比较）
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, char := range s {
		result[i] = unicode.ToLower(char)
	}
	return string(result)
}

// GetPasswordStrength 获取密码强度等级（0-4）
func GetPasswordStrength(password string) int {
	strength := 0

	// 长度评分
	if len(password) >= 8 {
		strength++
	}
	if len(password) >= 12 {
		strength++
	}

	// 复杂度评分
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasNumber = true
		} else if !unicode.IsSpace(char) {
			hasSpecial = true
		}
	}

	if hasUpper && hasLower {
		strength++
	}
	if hasNumber {
		strength++
	}
	if hasSpecial {
		strength++
	}

	// 限制最大强度
	if strength > 4 {
		strength = 4
	}

	return strength
}

// ValidatePasswordMatch 验证两次密码是否一致
func ValidatePasswordMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return cerr.BadRequest("password and confirmation password do not match")
	}
	return nil
}

// FormatPasswordRequirements 格式化密码要求（用于错误提示）
func FormatPasswordRequirements(config *PasswordStrengthConfig) string {
	if config == nil {
		config = DefaultPasswordConfig()
	}

	requirements := fmt.Sprintf("密码要求：长度 %d-%d 个字符", config.MinLength, config.MaxLength)

	if config.RequireUpper {
		requirements += "，包含大写字母"
	}
	if config.RequireLower {
		requirements += "，包含小写字母"
	}
	if config.RequireNumber {
		requirements += "，包含数字"
	}
	if config.RequireSpecial {
		requirements += "，包含特殊字符"
	}

	return requirements
}
