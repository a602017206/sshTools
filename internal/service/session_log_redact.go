package service

import "regexp"

var (
	sessionLogPasswordRe = regexp.MustCompile(`(?i)(password|passwd)\s*=\s*\S+`)
	sessionLogBearerRe   = regexp.MustCompile(`(?i)(bearer\s+)[a-z0-9._\-]+`)
	sessionLogPEMRe      = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
	sessionLogLongHexRe  = regexp.MustCompile(`\b[a-fA-F0-9]{32,}\b`)
	sessionLogLongB64Re  = regexp.MustCompile(`\b[A-Za-z0-9+/]{40,}={0,2}\b`)
)

// RedactSessionLog removes sensitive values from session log text before persistence.
func RedactSessionLog(text string) string {
	text = sessionLogPasswordRe.ReplaceAllString(text, "${1}=***")
	text = sessionLogBearerRe.ReplaceAllString(text, "${1}***")
	text = sessionLogPEMRe.ReplaceAllString(text, "***")
	text = sessionLogLongHexRe.ReplaceAllString(text, "***")
	text = sessionLogLongB64Re.ReplaceAllString(text, "***")
	return text
}
