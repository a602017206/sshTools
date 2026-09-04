package service

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const maxSQLStatementBytes = 32 * 1024 * 1024

type sqlSplitState int

const (
	sqlNormal sqlSplitState = iota
	sqlSingleQuote
	sqlDoubleQuote
	sqlLineComment
	sqlBlockComment
)

func SplitSQLScript(r io.Reader, dialect string) ([]string, error) {
	var statements []string
	err := EachSQLStatement(r, dialect, func(stmt string) error {
		statements = append(statements, stmt)
		return nil
	})
	return statements, err
}

func EachSQLStatement(r io.Reader, dialect string, fn func(string) error) error {
	if fn == nil {
		return fmt.Errorf("statement handler is required")
	}
	oracle := strings.EqualFold(strings.TrimSpace(dialect), "oracle")
	reader := bufio.NewReader(r)
	var stmt strings.Builder
	var line strings.Builder
	state := sqlNormal
	lineStart := true

	flush := func() error {
		text := strings.TrimSpace(stmt.String())
		stmt.Reset()
		if text == "" {
			return nil
		}
		return fn(text)
	}

	for {
		ch, _, err := reader.ReadRune()
		if err == io.EOF {
			if state == sqlBlockComment {
				return fmt.Errorf("SQL 文件有未闭合的块注释")
			}
			if oracle && strings.TrimSpace(line.String()) == "/" {
				line.Reset()
			} else {
				stmt.WriteString(line.String())
			}
			return flush()
		}
		if err != nil {
			return err
		}
		if stmt.Len()+line.Len() > maxSQLStatementBytes {
			return fmt.Errorf("单条 SQL 超过 %d 字节，无法执行", maxSQLStatementBytes)
		}

		switch state {
		case sqlLineComment:
			if ch == '\n' {
				state = sqlNormal
				lineStart = true
			}
			continue
		case sqlBlockComment:
			if ch == '*' {
				next, _, peekErr := reader.ReadRune()
				if peekErr == io.EOF {
					continue
				}
				if peekErr != nil {
					return peekErr
				}
				if next == '/' {
					state = sqlNormal
					continue
				}
				_ = reader.UnreadRune()
			}
			if ch == '\n' {
				lineStart = true
			}
			continue
		case sqlSingleQuote:
			line.WriteRune(ch)
			if ch == '\'' {
				next, _, peekErr := reader.ReadRune()
				if peekErr == nil && next == '\'' {
					line.WriteRune(next)
					continue
				}
				if peekErr == nil {
					_ = reader.UnreadRune()
				}
				state = sqlNormal
			}
			continue
		case sqlDoubleQuote:
			line.WriteRune(ch)
			if ch == '"' {
				next, _, peekErr := reader.ReadRune()
				if peekErr == nil && next == '"' {
					line.WriteRune(next)
					continue
				}
				if peekErr == nil {
					_ = reader.UnreadRune()
				}
				state = sqlNormal
			}
			continue
		}

		if lineStart && !unicode.IsSpace(ch) && ch != '-' && ch != '/' {
			lineStart = false
		}

		if state == sqlNormal && ch == '-' {
			next, _, peekErr := reader.ReadRune()
			if peekErr == nil && next == '-' {
				state = sqlLineComment
				lineStart = false
				continue
			}
			if peekErr == nil {
				_ = reader.UnreadRune()
			}
		}

		if state == sqlNormal && ch == '/' {
			next, _, peekErr := reader.ReadRune()
			if peekErr == nil && next == '*' {
				state = sqlBlockComment
				lineStart = false
				continue
			}
			if peekErr == nil {
				_ = reader.UnreadRune()
			}
			if oracle && lineStart && isOracleSlashTerminator(reader, ch) {
				if err := flush(); err != nil {
					return err
				}
				skipToNewline(reader)
				line.Reset()
				lineStart = true
				continue
			}
		}

		if state == sqlNormal && ch == '\'' {
			line.WriteRune(ch)
			state = sqlSingleQuote
			lineStart = false
			continue
		}
		if state == sqlNormal && ch == '"' {
			line.WriteRune(ch)
			state = sqlDoubleQuote
			lineStart = false
			continue
		}

		if state == sqlNormal && ch == ';' {
			stmt.WriteString(line.String())
			line.Reset()
			if err := flush(); err != nil {
				return err
			}
			lineStart = true
			continue
		}

		line.WriteRune(ch)
		if ch == '\n' {
			if oracle && strings.TrimSpace(line.String()) == "/" {
				line.Reset()
				if err := flush(); err != nil {
					return err
				}
				lineStart = true
				continue
			}
			stmt.WriteString(line.String())
			line.Reset()
			lineStart = true
		} else if !unicode.IsSpace(ch) {
			lineStart = false
		}
	}
}

func isOracleSlashTerminator(reader *bufio.Reader, slash rune) bool {
	_ = slash
	for {
		next, _, err := reader.ReadRune()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
		if next == '\n' || next == '\r' {
			_ = reader.UnreadRune()
			return true
		}
		if unicode.IsSpace(next) {
			continue
		}
		_ = reader.UnreadRune()
		return false
	}
}

func skipToNewline(reader *bufio.Reader) {
	for {
		ch, _, err := reader.ReadRune()
		if err != nil || ch == '\n' {
			return
		}
	}
}

func SQLFilePreamble(dbType, database, schema string) string {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "mysql":
		if strings.TrimSpace(database) == "" {
			return ""
		}
		return "USE " + quoteSQLIdent(database, '`')
	case "oracle":
		if strings.TrimSpace(schema) == "" {
			return ""
		}
		return "ALTER SESSION SET CURRENT_SCHEMA = " + quoteSQLIdent(schema, '"')
	case "postgresql", "kingbase", "opengauss":
		if strings.TrimSpace(schema) == "" {
			return ""
		}
		return "SET search_path TO " + quoteSQLIdent(schema, '"')
	default:
		if strings.TrimSpace(schema) != "" {
			return "ALTER SESSION SET CURRENT_SCHEMA = " + quoteSQLIdent(schema, '"')
		}
		if strings.TrimSpace(database) != "" {
			return "USE " + quoteSQLIdent(database, '"')
		}
		return ""
	}
}

func quoteSQLIdent(name string, quote rune) string {
	escaped := strings.ReplaceAll(name, string(quote), string(quote)+string(quote))
	return string(quote) + escaped + string(quote)
}
