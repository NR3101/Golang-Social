package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridMailer implements the mailer.Client interface using SendGrid for sending emails.
type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

// Constructs a new SendGridMailer with the provided fromEmail and API key.
func NewSendGridMailer(fromEmail, apiKey string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

// Send sends an email using the SendGrid API with the specified template, username, email, and data.
func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmp, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmp.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmp.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 0; i < maxRetries; i++ {
		respone, err := m.client.Send(message)
		if err != nil {
			log.Printf("failed to send email to: %v, attmept %d/%d: %v", email, i+1, maxRetries, err.Error())

			time.Sleep(time.Second * time.Duration(i+1)) // wait before retrying
			continue                                     // retry sending the email
		}

		log.Printf("email sent to: %s, status code: %d", email, respone.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email to %s after %d attempts", email, maxRetries)
}
