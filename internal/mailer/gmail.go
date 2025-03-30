package mailer

import (
	"bytes"
	"errors"
	gomail "gopkg.in/mail.v2"
	"html/template"
)

type gmailClient struct {
	fromEmail string
	password  string
}

func NewGmailClient(fromEmail, password string) (gmailClient, error) {
	if fromEmail == "" || password == "" {
		return gmailClient{}, errors.New("email and password are required")
	}

	return gmailClient{
		fromEmail: fromEmail,
		password:  password,
	}, nil
}

func (m *gmailClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
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
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())
	message.AddAlternative("text/html", body.String())

	// تنظیمات SMTP جیمیل
	dialer := gomail.NewDialer(
		"smtp.gmail.com", // سرور SMTP جیمیل
		587,              // پورت جیمیل
		m.fromEmail,      // آدرس ایمیل شما
		m.password,       // رمز عبور برنامه (نه رمز حساب جیمیل)
	)

	// این خط برای اتصال امن لازم است
	dialer.StartTLSPolicy = gomail.MandatoryStartTLS

	if err := dialer.DialAndSend(message); err != nil {
		return -1, err
	}

	return 200, nil
}
