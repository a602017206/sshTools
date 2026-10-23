package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	cryptossh "golang.org/x/crypto/ssh"
)

// sshDial is used by Connect; tests may stub slow dials.
var sshDial = cryptossh.Dial

// Client represents an SSH client
type Client struct {
	config  *cryptossh.ClientConfig
	client  *cryptossh.Client
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
	var authMethods []cryptossh.AuthMethod

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
		authMethods = append(authMethods, cryptossh.Password(cfg.Password))

		// Add keyboard-interactive authentication (some servers require this instead of password)
		authMethods = append(authMethods, cryptossh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
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

	config := &cryptossh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         cfg.Timeout,
	}

	return &Client{
		config:  config,
		address: address,
	}, nil
}

// Connect establishes SSH connection
func (c *Client) Connect() error {
	client, err := sshDial("tcp", c.address, c.config)
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

// Dial opens a connection to addr on the remote host over the SSH connection.
func (c *Client) Dial(network, addr string) (net.Conn, error) {
	if c.client == nil {
		return nil, fmt.Errorf("ssh client not connected")
	}
	return c.client.Dial(network, addr)
}

// CombinedOutput runs cmd in a new session on the remote host and returns
// its combined stdout and stderr. It does not allocate a PTY.
func (c *Client) CombinedOutput(cmd string) ([]byte, error) {
	if c.client == nil {
		return nil, fmt.Errorf("ssh client not connected")
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()
	return session.CombinedOutput(cmd)
}

// getKeyAuth loads a private key file and returns an SSH auth method
func getKeyAuth(keyPath, passphrase string) (cryptossh.AuthMethod, error) {
	// Read the private key file
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	var signer cryptossh.Signer

	// Try to parse the key with passphrase if provided
	if passphrase != "" {
		signer, err = cryptossh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse encrypted key (check passphrase): %w", err)
		}
	} else {
		// Try to parse without passphrase
		signer, err = cryptossh.ParsePrivateKey(keyBytes)
		if err != nil {
			// If it fails, it might be an encrypted key without passphrase provided
			if err.Error() == "ssh: cannot decode encrypted private keys" ||
				err.Error() == "ssh: this private key is passphrase protected" {
				return nil, fmt.Errorf("private key is encrypted, please provide passphrase")
			}
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}
	}

	return cryptossh.PublicKeys(signer), nil
}
