package mailer

import (
	"bytes"
	"errors"
	gomail "gopkg.in/mail.v2"
	"html/template"
)

type mailtrapClient struct {
	fromEmail string
	apiKey    string
}

func NewMailTrapClient(apikey, fromEmail string) (mailtrapClient, error) {
	if apikey == "" {
		return mailtrapClient{}, errors.New("api key is required")
	}

	return mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apikey,
	}, nil
}

func (m *mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	// template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}

	message := gomail.NewMessage()

	message.SetHeader("fROM", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())
	message.AddAlternative("text/html", body.String())

	//dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)
	dialer := gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, "50427fb3b0e8ad", "94bebafe1bc546")
	//dialer := gomail.NewDialer("sandbox.smtp.mailtrap.io", 587, "api", m.apiKey)

	if err := dialer.DialAndSend(message); err != nil {
		return -1, err
	}

	return 200, nil
}
