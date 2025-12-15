package ssh

import (
	"fmt"
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
	Host     string
	Port     int
	User     string
	Password string
	KeyPath  string
	Timeout  time.Duration
}

// NewClient creates a new SSH client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Build auth methods
	var authMethods []ssh.AuthMethod
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	// TODO: Add key-based authentication

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
