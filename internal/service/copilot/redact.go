package copilot

import "regexp"

var (
	passwordRe = regexp.MustCompile(`(?i)password\s*=\s*\S+`)
	pemKeyRe   = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
)

// Redact strips password assignments and PEM private key blocks.
func Redact(text string) string {
	text = passwordRe.ReplaceAllString(text, "password=[REDACTED]")
	text = pemKeyRe.ReplaceAllString(text, "[REDACTED PRIVATE KEY]")
	return text
}
