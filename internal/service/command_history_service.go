package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const commandHistoryMaxEntries = 500

// CommandHistoryEntry is a single command usage record returned by Suggest.
type CommandHistoryEntry struct {
	Command  string
	Count    int
	LastUsed time.Time
}

type commandHistoryEntryStored struct {
	Command  string    `json:"command"`
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

type commandHistoryFile struct {
	ConnectionID string                      `json:"connection_id"`
	UpdatedAt    time.Time                   `json:"updated_at"`
	Entries      []commandHistoryEntryStored `json:"entries"`
}

// CommandHistoryService persists per-connection command usage for suggestions.
type CommandHistoryService struct {
	rootDir string
	mu      sync.Mutex
}

// NewCommandHistoryService creates a command history service rooted at rootDir.
func NewCommandHistoryService(rootDir string) *CommandHistoryService {
	return &CommandHistoryService{rootDir: rootDir}
}

func (s *CommandHistoryService) filePath(connectionID string) string {
	return filepath.Join(s.rootDir, connectionID+".json")
}

// Record increments usage for command on connectionID. Empty commands are ignored.
func (s *CommandHistoryService) Record(connectionID, command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.loadLocked(connectionID)
	if err != nil {
		return err
	}
	if file == nil {
		file = &commandHistoryFile{ConnectionID: connectionID}
	}

	now := time.Now().UTC()
	found := false
	for i := range file.Entries {
		if file.Entries[i].Command == command {
			file.Entries[i].Count++
			file.Entries[i].LastUsed = now
			found = true
			break
		}
	}
	if !found {
		file.Entries = append(file.Entries, commandHistoryEntryStored{
			Command:  command,
			Count:    1,
			LastUsed: now,
		})
	}

	if len(file.Entries) > commandHistoryMaxEntries {
		sort.Slice(file.Entries, func(i, j int) bool {
			return file.Entries[i].LastUsed.Before(file.Entries[j].LastUsed)
		})
		file.Entries = file.Entries[len(file.Entries)-commandHistoryMaxEntries:]
	}

	file.UpdatedAt = now
	return s.saveLocked(connectionID, file)
}

// Suggest returns up to limit matching commands for connectionID.
// Prefix matches (HasPrefix) are returned before substring matches (Contains).
func (s *CommandHistoryService) Suggest(connectionID, prefix string, limit int) ([]CommandHistoryEntry, error) {
	if limit <= 0 {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := s.loadLocked(connectionID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, nil
	}

	var prefixMatches, containsMatches []CommandHistoryEntry
	for _, entry := range file.Entries {
		item := CommandHistoryEntry{
			Command:  entry.Command,
			Count:    entry.Count,
			LastUsed: entry.LastUsed,
		}
		if strings.HasPrefix(entry.Command, prefix) {
			prefixMatches = append(prefixMatches, item)
		} else if strings.Contains(entry.Command, prefix) {
			containsMatches = append(containsMatches, item)
		}
	}

	sortCommandHistoryEntries(prefixMatches)
	sortCommandHistoryEntries(containsMatches)

	result := prefixMatches
	if len(result) < limit {
		need := limit - len(result)
		if need > len(containsMatches) {
			need = len(containsMatches)
		}
		result = append(result, containsMatches[:need]...)
	} else {
		result = result[:limit]
	}
	return result, nil
}

func sortCommandHistoryEntries(entries []CommandHistoryEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].LastUsed.After(entries[j].LastUsed)
	})
}

func (s *CommandHistoryService) loadLocked(connectionID string) (*commandHistoryFile, error) {
	data, err := os.ReadFile(s.filePath(connectionID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("command history: read: %w", err)
	}

	var file commandHistoryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("command history: decode: %w", err)
	}
	return &file, nil
}

func (s *CommandHistoryService) saveLocked(connectionID string, file *commandHistoryFile) error {
	if err := os.MkdirAll(s.rootDir, 0o755); err != nil {
		return fmt.Errorf("command history: create directory: %w", err)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("command history: encode: %w", err)
	}

	path := s.filePath(connectionID)
	temp, err := os.CreateTemp(s.rootDir, ".command-history-*.tmp")
	if err != nil {
		return fmt.Errorf("command history: create temp: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("command history: write temp: %w", err)
	}
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return fmt.Errorf("command history: chmod temp: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("command history: close temp: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("command history: rename: %w", err)
	}
	return nil
}
