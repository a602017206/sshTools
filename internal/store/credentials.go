package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// CredentialStore manages SSH credentials with encryption
type CredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]string // connectionID -> encrypted password
	storePath   string
	key         []byte // Encryption key derived from machine ID
}

// NewCredentialStore creates a new credential store
func NewCredentialStore() *CredentialStore {
	homeDir, _ := os.UserHomeDir()
	storePath := filepath.Join(homeDir, ".sshtools", "credentials.enc")

	// Generate encryption key from machine-specific data
	// In production, you might want to use system keychain instead
	key := deriveEncryptionKey()

	store := &CredentialStore{
		credentials: make(map[string]string),
		storePath:   storePath,
		key:         key,
	}

	// Load existing credentials
	_ = store.load()

	return store
}

// deriveEncryptionKey generates an encryption key
// In production, this should use system keychain or user password
func deriveEncryptionKey() []byte {
	// Use a combination of machine-specific data
	// This is a basic implementation - for production use system keychain
	hostname, _ := os.Hostname()
	homeDir, _ := os.UserHomeDir()

	data := fmt.Sprintf("%s:%s", hostname, homeDir)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// Store stores a password for a connection
func (s *CredentialStore) Store(connectionID, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Encrypt password before storing
	encrypted, err := s.encrypt(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	s.credentials[connectionID] = encrypted

	// Persist to disk
	return s.save()
}

// Get retrieves a password for a connection
func (s *CredentialStore) Get(connectionID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encrypted, exists := s.credentials[connectionID]
	if !exists {
		return "", fmt.Errorf("credential not found for connection: %s", connectionID)
	}

	// Decrypt password before returning
	password, err := s.decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return password, nil
}

// Delete removes a stored password
func (s *CredentialStore) Delete(connectionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.credentials, connectionID)

	// Persist to disk
	return s.save()
}

// Has checks if a password exists for a connection
func (s *CredentialStore) Has(connectionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.credentials[connectionID]
	return exists
}

// encrypt encrypts a password using AES-GCM
func (s *CredentialStore) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a password using AES-GCM
func (s *CredentialStore) decrypt(encryptedText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// save persists credentials to disk
func (s *CredentialStore) save() error {
	data, err := json.MarshalIndent(s.credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.storePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(s.storePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// load loads credentials from disk
func (s *CredentialStore) load() error {
	data, err := os.ReadFile(s.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's okay
			return nil
		}
		return fmt.Errorf("failed to read credentials: %w", err)
	}

	if err := json.Unmarshal(data, &s.credentials); err != nil {
		return fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return nil
}
