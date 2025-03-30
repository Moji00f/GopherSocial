package mailer

import "embed"

//go:embed "templates"
var FS embed.FS

const (
	FromName            = "GoDevOps"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) (int, error)
}
