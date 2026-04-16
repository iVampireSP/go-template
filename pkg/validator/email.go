package validator

import (
	"strings"

	"github.com/iVampireSP/go-template/pkg/cerr"
)

// 临时邮箱域名黑名单（常见的一次性邮箱服务）
var blockedEmailDomains = map[string]bool{
	// 一次性邮箱服务
	"tempmail.com":          true,
	"10minutemail.com":      true,
	"guerrillamail.com":     true,
	"mailinator.com":        true,
	"throwaway.email":       true,
	"temp-mail.org":         true,
	"getnada.com":           true,
	"maildrop.cc":           true,
	"trashmail.com":         true,
	"fakeinbox.com":         true,
	"yopmail.com":           true,
	"sharklasers.com":       true,
	"grr.la":                true,
	"guerrillamail.biz":     true,
	"guerrillamail.de":      true,
	"spam4.me":              true,
	"mytemp.email":          true,
	"tempinbox.com":         true,
	"mohmal.com":            true,
	"emailondeck.com":       true,
	"minute-mail.com":       true,
	"dispostable.com":       true,
	"mintemail.com":         true,
	"mailnesia.com":         true,
	"emailfake.com":         true,
	"throwawaymail.com":     true,
	"fakemailgenerator.com": true,
	"generator.email":       true,
	"inbox.lv":              true,
	"mailcatch.com":         true,
	"mailsac.com":           true,
	"mvrht.com":             true,
	"spambox.us":            true,
	"tempail.com":           true,
	"jetable.org":           true,
	"33mail.com":            true,
	"mailbox.in.ua":         true,
	"tmails.net":            true,
	"crazymailing.com":      true,
	"bccto.me":              true,
	"tmail.ws":              true,
}

var (
	ErrEmailDomainBlocked = cerr.BadRequest("email domain is not allowed for registration, please use a trusted email provider")
)

// ValidateEmailDomain 验证邮箱域名是否在黑名单中
func ValidateEmailDomain(email string) error {
	// 分离域名
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return cerr.BadRequest("invalid email format")
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))

	// 检查是否在黑名单中
	if blockedEmailDomains[domain] {
		return ErrEmailDomainBlocked
	}

	return nil
}

// IsEmailDomainBlocked 检查邮箱域名是否被封禁（返回布尔值）
func IsEmailDomainBlocked(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	return blockedEmailDomains[domain]
}

// AddBlockedDomain 添加黑名单域名（用于动态配置）
func AddBlockedDomain(domain string) {
	blockedEmailDomains[strings.ToLower(strings.TrimSpace(domain))] = true
}

// RemoveBlockedDomain 移除黑名单域名（用于动态配置）
func RemoveBlockedDomain(domain string) {
	delete(blockedEmailDomains, strings.ToLower(strings.TrimSpace(domain)))
}
