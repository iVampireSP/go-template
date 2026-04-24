package email

import (
	"github.com/iVampireSP/go-template/internal/infra/config"
	foundationemail "github.com/iVampireSP/go-template/pkg/foundation/email"
)

// Re-export foundation types.
type Email = foundationemail.Email
type Content = foundationemail.Content
type Mailable = foundationemail.Mailable

func NewEmail() (*Email, error) {
	return foundationemail.NewEmail(foundationemail.Config{
		Host:       config.String("mail.host"),
		Port:       config.Int("mail.port", 587),
		Username:   config.String("mail.username"),
		Password:   config.String("mail.password"),
		From:       config.String("mail.from"),
		Encryption: config.String("mail.encryption", "tls"),
	}), nil
}
