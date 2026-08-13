package copilot

import "regexp"

var (
	sqlDrop     = regexp.MustCompile(`(?i)\bDROP\b`)
	sqlDelete   = regexp.MustCompile(`(?i)\bDELETE\b`)
	sqlTruncate = regexp.MustCompile(`(?i)\bTRUNCATE\b`)
	sqlUpdate   = regexp.MustCompile(`(?i)\bUPDATE\b`)
	sqlWhere    = regexp.MustCompile(`(?i)\bWHERE\b`)

	shellPatterns = []struct {
		re     *regexp.Regexp
		reason string
	}{
		{regexp.MustCompile(`(?i)\brm\b`), "rm"},
		{regexp.MustCompile(`(?i)\bmkfs\b`), "mkfs"},
		{regexp.MustCompile(`(?i)\bdd\b`), "dd"},
		{regexp.MustCompile(`(?i)\bshutdown\b`), "shutdown"},
		{regexp.MustCompile(`(?i)\breboot\b`), "reboot"},
		{regexp.MustCompile(`(?i)\bkill\s+-9\b`), "kill -9"},
		{regexp.MustCompile(`(?i)\bchmod\s+777\b`), "chmod 777"},
		{regexp.MustCompile(`>\s*/dev/sd`), ">/dev/sd"},
	}
)

// Classify marks SQL/shell content as destructive using local rules.
func Classify(kind, content string) Result {
	switch kind {
	case "sql":
		return classifySQL(content)
	case "shell":
		return classifyShell(content)
	default:
		return Result{}
	}
}

func classifySQL(content string) Result {
	switch {
	case sqlDrop.MatchString(content):
		return Result{Destructive: true, Reason: "DROP"}
	case sqlDelete.MatchString(content):
		return Result{Destructive: true, Reason: "DELETE"}
	case sqlTruncate.MatchString(content):
		return Result{Destructive: true, Reason: "TRUNCATE"}
	case sqlUpdate.MatchString(content) && !sqlWhere.MatchString(content):
		return Result{Destructive: true, Reason: "UPDATE without WHERE"}
	default:
		return Result{}
	}
}

func classifyShell(content string) Result {
	for _, p := range shellPatterns {
		if p.re.MatchString(content) {
			return Result{Destructive: true, Reason: p.reason}
		}
	}
	return Result{}
}
