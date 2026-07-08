package ssh

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client
type Client struct {
	config  *ssh.ClientConfig
	client  *ssh.Client
	address string
}

// Config holds SSH connection configuration
type Config struct {
	Host       string
	Port       int
	User       string
	Password   string
	KeyPath    string
	Passphrase string // Passphrase for encrypted private keys
	Timeout    time.Duration
}

// NewClient creates a new SSH client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Build auth methods
	var authMethods []ssh.AuthMethod

	// Try key-based authentication first (if key path is provided)
	if cfg.KeyPath != "" {
		keyAuth, err := getKeyAuth(cfg.KeyPath, cfg.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to load SSH key: %w", err)
		}
		authMethods = append(authMethods, keyAuth)
	}

	// Add password authentication
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))

		// Add keyboard-interactive authentication (some servers require this instead of password)
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range questions {
				answers[i] = cfg.Password
			}
			return answers, nil
		}))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (need password or key path)")
	}

	config := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         cfg.Timeout,
	}

	return &Client{
		config:  config,
		address: address,
	}, nil
}

// Connect establishes SSH connection
func (c *Client) Connect() error {
	client, err := ssh.Dial("tcp", c.address, c.config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.client = client
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsConnected checks if client is connected
func (c *Client) IsConnected() bool {
	return c.client != nil
}

// getKeyAuth loads a private key file and returns an SSH auth method
func getKeyAuth(keyPath, passphrase string) (ssh.AuthMethod, error) {
	// Read the private key file
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	var signer ssh.Signer

	// Try to parse the key with passphrase if provided
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse encrypted key (check passphrase): %w", err)
		}
	} else {
		// Try to parse without passphrase
		signer, err = ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			// If it fails, it might be an encrypted key without passphrase provided
			if err.Error() == "ssh: cannot decode encrypted private keys" ||
				err.Error() == "ssh: this private key is passphrase protected" {
				return nil, fmt.Errorf("private key is encrypted, please provide passphrase")
			}
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}
	}

	return ssh.PublicKeys(signer), nil
}
