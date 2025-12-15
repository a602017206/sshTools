package store

import (
	"fmt"
	"sync"
)

// CredentialStore manages SSH credentials
// TODO: Implement secure storage using system keychain
type CredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]string // connectionID -> password
}

// NewCredentialStore creates a new credential store
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		credentials: make(map[string]string),
	}
}

// Store stores a password for a connection
func (s *CredentialStore) Store(connectionID, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Encrypt password before storing
	s.credentials[connectionID] = password
	return nil
}

// Get retrieves a password for a connection
func (s *CredentialStore) Get(connectionID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	password, exists := s.credentials[connectionID]
	if !exists {
		return "", fmt.Errorf("credential not found for connection: %s", connectionID)
	}

	// TODO: Decrypt password before returning
	return password, nil
}

// Delete removes a stored password
func (s *CredentialStore) Delete(connectionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.credentials, connectionID)
	return nil
}

// Has checks if a password exists for a connection
func (s *CredentialStore) Has(connectionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.credentials[connectionID]
	return exists
}
