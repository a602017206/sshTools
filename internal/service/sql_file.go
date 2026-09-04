package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	sqlFileStatementTimeout = 5 * time.Minute
	sqlFileFailedSQLLimit   = 500
	sqlFileProgressEvery    = 20
)

type SQLFileRequest struct {
	SessionID string
	Path      string
	Database  string
	Schema    string
}

type SQLFileProgress struct {
	SessionID  string `json:"sessionId"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	BytesRead  int64  `json:"bytesRead"`
	Statements int    `json:"statements"`
	Affected   int    `json:"affected"`
	Done       bool   `json:"done"`
	Canceled   bool   `json:"canceled"`
	Error      string `json:"error"`
	FailedSQL  string `json:"failedSql"`
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func (ds *DatabaseService) ExecuteSQLFile(ctx context.Context, req SQLFileRequest, progress func(SQLFileProgress)) error {
	current := SQLFileProgress{
		SessionID: req.SessionID,
		FileName:  filepath.Base(req.Path),
	}
	emit := func() {
		if progress != nil {
			progress(current)
		}
	}
	finish := func(execErr error) error {
		current.Done = true
		if execErr != nil {
			if errors.Is(execErr, context.Canceled) {
				current.Canceled = true
				current.Error = "已取消"
			} else {
				current.Error = execErr.Error()
			}
		}
		emit()
		return execErr
	}

	if strings.TrimSpace(req.Path) == "" {
		return finish(fmt.Errorf("SQL 文件路径不能为空"))
	}
	session, err := ds.GetSession(req.SessionID)
	if err != nil {
		return finish(err)
	}

	file, err := os.Open(req.Path)
	if err != nil {
		return finish(fmt.Errorf("无法打开 SQL 文件: %w", err))
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return finish(fmt.Errorf("无法读取 SQL 文件信息: %w", err))
	}
	current.FileSize = info.Size()
	emit()

	counter := &countingReader{r: file}
	trackedEmit := func() {
		current.BytesRead = counter.n
		emit()
	}

	dialect := session.Config.DBType
	if preamble := SQLFilePreamble(dialect, req.Database, req.Schema); preamble != "" {
		if _, err := ds.executeSQLFileStatement(ctx, req.SessionID, preamble); err != nil {
			current.FailedSQL = truncateSQLPreview(preamble)
			return finish(err)
		}
		trackedEmit()
	}

	err = EachSQLStatement(counter, dialect, func(stmt string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, execErr := ds.executeSQLFileStatement(ctx, req.SessionID, stmt)
		if execErr != nil {
			current.FailedSQL = truncateSQLPreview(stmt)
			return execErr
		}
		current.Statements++
		if result != nil {
			current.Affected += result.Affected
		}
		if current.Statements%sqlFileProgressEvery == 0 {
			trackedEmit()
		}
		return nil
	})
	current.BytesRead = counter.n
	return finish(err)
}

func (ds *DatabaseService) executeSQLFileStatement(ctx context.Context, sessionID, query string) (*QueryResult, error) {
	stmtCtx, cancel := context.WithTimeout(ctx, sqlFileStatementTimeout)
	defer cancel()
	if ds.gateway != nil {
		return ds.gateway.ExecuteQuery(stmtCtx, sessionID, query)
	}
	_ = stmtCtx
	return ds.ExecuteQuery(sessionID, query)
}

func truncateSQLPreview(sql string) string {
	text := strings.TrimSpace(sql)
	if utf8.RuneCountInString(text) <= sqlFileFailedSQLLimit {
		return text
	}
	return string([]rune(text)[:sqlFileFailedSQLLimit]) + "…"
}
