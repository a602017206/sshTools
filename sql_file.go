package main

import (
	"context"
	"fmt"
	"path/filepath"

	"AHaSSHTools/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const sqlFileProgressEvent = "sqlfile:progress"

func (a *App) SelectSQLFile() (string, error) {
	if a == nil || a.ctx == nil {
		return "", fmt.Errorf("应用未初始化")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 SQL 文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQL 文件 (*.sql)", Pattern: "*.sql"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
}

func (a *App) StartSQLFile(sessionID, filePath, database, schema string) error {
	if a == nil || a.databaseService == nil {
		return fmt.Errorf("数据库服务未初始化")
	}
	if sessionID == "" || filePath == "" {
		return fmt.Errorf("会话或文件路径不能为空")
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.sqlFileMu.Lock()
	if a.sqlFileCancel == nil {
		a.sqlFileCancel = map[string]context.CancelFunc{}
	}
	if _, busy := a.sqlFileCancel[sessionID]; busy {
		a.sqlFileMu.Unlock()
		cancel()
		return fmt.Errorf("该会话已有 SQL 文件正在执行")
	}
	a.sqlFileCancel[sessionID] = cancel
	a.sqlFileMu.Unlock()

	go func() {
		defer func() {
			a.sqlFileMu.Lock()
			delete(a.sqlFileCancel, sessionID)
			a.sqlFileMu.Unlock()
			cancel()
		}()
		_ = a.databaseService.ExecuteSQLFile(ctx, service.SQLFileRequest{
			SessionID: sessionID,
			Path:      filePath,
			Database:  database,
			Schema:    schema,
		}, func(progress service.SQLFileProgress) {
			if progress.SessionID == "" {
				progress.SessionID = sessionID
			}
			if progress.FileName == "" {
				progress.FileName = filepath.Base(filePath)
			}
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, sqlFileProgressEvent, progress)
			}
		})
	}()
	return nil
}

func (a *App) CancelSQLFile(sessionID string) {
	if a == nil || sessionID == "" {
		return
	}
	a.sqlFileMu.Lock()
	defer a.sqlFileMu.Unlock()
	if cancel := a.sqlFileCancel[sessionID]; cancel != nil {
		cancel()
	}
}
