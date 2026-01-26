package mailer

import (
	"context"
	"errors"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type MailService struct {
	mg *mailgun.MailgunImpl
}

func NewMailService() (*MailService, error) {
	apiKey := "3b692fc3adf4394dcedc3fffd323183b-e61ae8dd-e3cdc048"
	domain := "oryoo.in"

	if apiKey == "" || domain == "" {
		return nil, errors.New("MAILGUN_API_KEY or MAILGUN_DOMAIN not set")
	}

	mg := mailgun.NewMailgun(domain, apiKey)

	return &MailService{mg: mg}, nil
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
