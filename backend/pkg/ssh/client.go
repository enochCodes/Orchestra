package ssh

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// CommandResult holds the output of a remote command execution.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Client wraps an SSH connection to a remote server.
type Client struct {
	host   string
	port   int
	user   string
	client *ssh.Client
}

// NormalizePEMKey fixes common PEM key formatting issues (extra line breaks, wrong wraps).
func NormalizePEMKey(key []byte) []byte {
	s := string(key)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	var out []string
	var base64Buf strings.Builder
	inBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "-----BEGIN ") {
			inBlock = true
			out = append(out, line)
			base64Buf.Reset()
			continue
		}
		if strings.HasPrefix(line, "-----END ") {
			if base64Buf.Len() > 0 {
				b64 := base64Buf.String()
				for len(b64) > 64 {
					out = append(out, b64[:64])
					b64 = b64[64:]
				}
				if len(b64) > 0 {
					out = append(out, b64)
				}
			}
			out = append(out, line)
			inBlock = false
			continue
		}
		if inBlock {
			base64Buf.WriteString(line)
		}
	}
	return []byte(strings.Join(out, "\n"))
}

// NewClient creates a new SSH client connection using a private key.
func NewClient(host string, port int, user string, privateKey []byte, passphrase string) (*Client, error) {
	normalized := NormalizePEMKey(privateKey)
	var signer ssh.Signer
	var err error

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(normalized, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(normalized)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: implement known_hosts verification
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return &Client{
		host:   host,
		port:   port,
		user:   user,
		client: client,
	}, nil
}

// NewClientWithPassword creates a new SSH client connection using password auth.
func NewClientWithPassword(host string, port int, user, password string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return &Client{
		host:   host,
		port:   port,
		user:   user,
		client: client,
	}, nil
}

// ExecuteCommand runs a command on the remote server and returns the result.
func (c *Client) ExecuteCommand(cmd string) (*CommandResult, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	result := &CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			return result, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// Close terminates the SSH connection.
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
