package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const sessionLogTimeLayout = "2006-01-02T15-04-05"

// SessionLogInfo describes a persisted session log file.
type SessionLogInfo struct {
	ID           string
	ConnectionID string
	SessionID    string
	Path         string
	Size         int64
	ModTime      time.Time
}

// SessionLogHit is a single search match within a session log.
type SessionLogHit struct {
	LogID string
	Line  int
	Text  string
}

type sessionLogWriter struct {
	file  *os.File
	logID string
}

// SessionLogService persists SSH session output to disk.
type SessionLogService struct {
	rootDir string
	mu      sync.Mutex
	writers map[string]*sessionLogWriter // key: connectionID/sessionID
}

// NewSessionLogService creates a session log service rooted at rootDir.
func NewSessionLogService(rootDir string) *SessionLogService {
	return &SessionLogService{
		rootDir: rootDir,
		writers: make(map[string]*sessionLogWriter),
	}
}

func (s *SessionLogService) writerKey(connectionID, sessionID string) string {
	return connectionID + "/" + sessionID
}

// Append writes data to the session log, creating the file on first write.
func (s *SessionLogService) Append(connectionID, sessionID string, data []byte, redact bool) error {
	if redact {
		data = []byte(RedactSessionLog(string(data)))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.writerKey(connectionID, sessionID)
	w, ok := s.writers[key]
	if !ok {
		connDir := filepath.Join(s.rootDir, connectionID)
		if err := os.MkdirAll(connDir, 0o700); err != nil {
			return fmt.Errorf("session log: create directory: %w", err)
		}

		name := time.Now().Format(sessionLogTimeLayout) + "_" + sessionID + ".log"
		path := filepath.Join(connDir, name)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("session log: open file: %w", err)
		}

		logID := filepath.ToSlash(filepath.Join(connectionID, name))
		w = &sessionLogWriter{file: f, logID: logID}
		s.writers[key] = w
	}

	if _, err := w.file.Write(data); err != nil {
		return fmt.Errorf("session log: write: %w", err)
	}
	return nil
}

// CloseSession closes the open writer for a session and removes it from the map.
// Missing writers are ignored.
func (s *SessionLogService) CloseSession(connectionID, sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.writerKey(connectionID, sessionID)
	w, ok := s.writers[key]
	if !ok {
		return
	}
	_ = w.file.Close()
	delete(s.writers, key)
}

// List returns session logs for a connection, newest first.
func (s *SessionLogService) List(connectionID string) ([]SessionLogInfo, error) {
	connDir := filepath.Join(s.rootDir, connectionID)
	entries, err := os.ReadDir(connDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("session log: list: %w", err)
	}

	var logs []SessionLogInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		info, err := s.buildLogInfo(connectionID, entry.Name())
		if err != nil {
			continue
		}
		logs = append(logs, info)
	}

	for i := 0; i < len(logs); i++ {
		for j := i + 1; j < len(logs); j++ {
			if logs[j].ModTime.After(logs[i].ModTime) {
				logs[i], logs[j] = logs[j], logs[i]
			}
		}
	}
	return logs, nil
}

func (s *SessionLogService) buildLogInfo(connectionID, filename string) (SessionLogInfo, error) {
	path := filepath.Join(s.rootDir, connectionID, filename)
	stat, err := os.Stat(path)
	if err != nil {
		return SessionLogInfo{}, err
	}

	sessionID := parseSessionIDFromLogFilename(filename)
	logID := filepath.ToSlash(filepath.Join(connectionID, filename))

	return SessionLogInfo{
		ID:           logID,
		ConnectionID: connectionID,
		SessionID:    sessionID,
		Path:         path,
		Size:         stat.Size(),
		ModTime:      stat.ModTime(),
	}, nil
}

func parseSessionIDFromLogFilename(filename string) string {
	base := strings.TrimSuffix(filename, ".log")
	if len(base) <= len(sessionLogTimeLayout)+1 {
		return base
	}
	return base[len(sessionLogTimeLayout)+1:]
}

func parseLogTimeFromFilename(filename string) (time.Time, bool) {
	base := strings.TrimSuffix(filename, ".log")
	if len(base) < len(sessionLogTimeLayout) {
		return time.Time{}, false
	}
	t, err := time.Parse(sessionLogTimeLayout, base[:len(sessionLogTimeLayout)])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// Search scans log files for query and returns up to limit hits.
func (s *SessionLogService) Search(connectionID, query string, limit int) ([]SessionLogHit, error) {
	if limit <= 0 {
		return nil, nil
	}
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	logs, err := s.List(connectionID)
	if err != nil {
		return nil, err
	}

	var hits []SessionLogHit
	for _, log := range logs {
		fileHits, err := searchLogFile(log.ID, log.Path, query)
		if err != nil {
			return nil, err
		}
		hits = append(hits, fileHits...)
		if len(hits) >= limit {
			return hits[:limit], nil
		}
	}
	return hits, nil
}

func searchLogFile(logID, path, query string) ([]SessionLogHit, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("session log: search open: %w", err)
	}
	defer f.Close()

	var hits []SessionLogHit
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		if strings.Contains(text, query) {
			hits = append(hits, SessionLogHit{
				LogID: logID,
				Line:  lineNum,
				Text:  text,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("session log: search read: %w", err)
	}
	return hits, nil
}

// Export copies a session log to destPath.
func (s *SessionLogService) Export(logID, destPath string) error {
	srcPath, err := s.resolveLogPath(logID)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("session log: export read: %w", err)
	}
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("session log: export write: %w", err)
	}
	return nil
}

// Delete removes a session log file.
func (s *SessionLogService) Delete(logID string) error {
	path, err := s.resolveLogPath(logID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, w := range s.writers {
		if w.logID == logID {
			_ = w.file.Close()
			delete(s.writers, key)
			break
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("session log: delete: %w", err)
	}
	return nil
}

// PurgeExpired deletes log files older than retentionDays and returns the count removed.
func (s *SessionLogService) PurgeExpired(retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	removed := 0

	connEntries, err := os.ReadDir(s.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("session log: purge list root: %w", err)
	}

	for _, connEntry := range connEntries {
		if !connEntry.IsDir() {
			continue
		}
		connID := connEntry.Name()
		logDir := filepath.Join(s.rootDir, connID)
		logEntries, err := os.ReadDir(logDir)
		if err != nil {
			continue
		}
		for _, logEntry := range logEntries {
			if logEntry.IsDir() || !strings.HasSuffix(logEntry.Name(), ".log") {
				continue
			}
			path := filepath.Join(logDir, logEntry.Name())
			fileTime, ok := parseLogTimeFromFilename(logEntry.Name())
			if !ok {
				stat, err := os.Stat(path)
				if err != nil {
					continue
				}
				fileTime = stat.ModTime()
			}
			if fileTime.After(cutoff) {
				continue
			}
			logID := filepath.ToSlash(filepath.Join(connID, logEntry.Name()))
			if err := s.Delete(logID); err != nil {
				return removed, err
			}
			removed++
		}
	}
	return removed, nil
}

func (s *SessionLogService) resolveLogPath(logID string) (string, error) {
	clean := filepath.Clean(logID)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("session log: invalid log id")
	}
	path := filepath.Join(s.rootDir, clean)
	absRoot, err := filepath.Abs(s.rootDir)
	if err != nil {
		return "", fmt.Errorf("session log: resolve root: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("session log: resolve path: %w", err)
	}
	if !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) && absPath != absRoot {
		return "", fmt.Errorf("session log: invalid log id")
	}
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("session log: stat: %w", err)
	}
	return absPath, nil
}
