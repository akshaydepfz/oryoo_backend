package mailer

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

const (
	mailgunDomain = "sandbox60dbb62bd7ab4ff5983193ef5186e0e0.mailgun.org"
	// DefaultFrom authorized send address.
	DefaultFrom = "Oryoo <hi@oryoo.in>"
	// AlertTo recipient for internal registration alerts.
	AlertTo = "hi@oryoo.in"
)

var (
	mailgunAPIKey = os.Getenv("MAILGUN_API_KEY")
)

type MailService struct {
	mg *mailgun.MailgunImpl
}

var (
	mailServiceInstance *MailService
	mailServiceOnce     sync.Once
)

func NewMailService() *MailService {
	mg := mailgun.NewMailgun(mailgunDomain, mailgunAPIKey)
	return &MailService{mg: mg}
}

// GetMailService returns a singleton MailService instance.
func GetMailService() *MailService {
	mailServiceOnce.Do(func() {
		mailServiceInstance = NewMailService()
	})
	return mailServiceInstance
}

func (m *MailService) SendMessage(from, subject, body, to string) (string, string, error) {
	message := m.mg.NewMessage(from, subject, body, to)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	return m.mg.Send(ctx, message)
}

func (m *MailService) SendWelcomeEmail(toEmail, toName string) error {
	message := m.mg.NewMessage(
		"Oryoo <noreply@oryoo.in>",
		"",
		"",
	)

	message.SetTemplate("welcome mail")

	message.AddRecipient(toName + " <" + toEmail + ">")

	message.AddTemplateVariable("name", toName)
	message.AddTemplateVariable("product", "Oryoo")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _, err := m.mg.Send(ctx, message)
	return err
}

// SendTestEmail sends a plain-text message (works on sandbox without Mailgun templates).
func (m *MailService) SendTestEmail(from, to, subject string) error {
	body := "This is a test message from the Oryoo backend (Mailgun)."
	_, _, err := m.SendMessage(from, subject, body, to)
	return err
}
