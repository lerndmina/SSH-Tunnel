package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// KeyManager handles SSH key operations
type KeyManager struct{}

// NewKeyManager creates a new SSH key manager
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// GenerateKeyPair generates a new SSH key pair
func (km *KeyManager) GenerateKeyPair(keyType, keyPath string) error {
	switch keyType {
	case "ed25519", "":
		return km.generateED25519KeyPair(keyPath)
	case "rsa":
		return fmt.Errorf("RSA key generation not yet implemented")
	case "ecdsa":
		return fmt.Errorf("ECDSA key generation not yet implemented")
	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// generateED25519KeyPair generates an ED25519 key pair
func (km *KeyManager) generateED25519KeyPair(keyPath string) error {
	// Generate ED25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ED25519 key pair: %w", err)
	}

	// Convert to SSH format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH public key: %w", err)
	}

	// Create private key PEM block
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	// Ensure directory exists
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write private key
	if err := os.WriteFile(keyPath, privPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key
	pubKeyPath := keyPath + ".pub"
	pubKeyData := ssh.MarshalAuthorizedKey(sshPubKey)
	if err := os.WriteFile(pubKeyPath, pubKeyData, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// ValidateKey validates an SSH private key
func (km *KeyManager) ValidateKey(keyPath string) error {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Try to parse as SSH private key
	_, err = ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("invalid SSH private key: %w", err)
	}

	return nil
}

// GetFingerprint gets the SSH fingerprint of a host
func (km *KeyManager) GetFingerprint(host string, port int) (string, error) {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Set timeout for connection
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	var hostKey ssh.PublicKey
	// Perform SSH handshake to get host key
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, &ssh.ClientConfig{
		User: "dummy", // We don't need to authenticate
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			hostKey = key
			return nil
		},
		Timeout: 5 * time.Second,
	})
	if err != nil && hostKey == nil {
		// Try to extract host key from error if possible
		return "", fmt.Errorf("failed SSH handshake with %s: %w", address, err)
	}
	if sshConn != nil {
		defer sshConn.Close()
		defer func() {
			go ssh.DiscardRequests(reqs)
			go func() {
				for newChannel := range chans {
					_ = newChannel.Reject(ssh.UnknownChannelType, "not implemented")
				}
			}()
		}()
	}

	// Get host key fingerprint
	if hostKey == nil {
		return "", fmt.Errorf("could not retrieve host key")
	}

	return ssh.FingerprintSHA256(hostKey), nil
}

// InstallPublicKey installs a public key on a remote server
func (km *KeyManager) InstallPublicKey(host, user, keyPath string, port int) error {
	// Read public key
	pubKeyPath := keyPath + ".pub"
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	// Read private key for authentication
	privKeyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(privKeyData)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to remote server
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Command to add public key to authorized_keys
	cmd := fmt.Sprintf(`
		mkdir -p ~/.ssh &&
		chmod 700 ~/.ssh &&
		echo '%s' >> ~/.ssh/authorized_keys &&
		chmod 600 ~/.ssh/authorized_keys &&
		echo "Public key installed successfully"
	`, string(pubKeyData))

	// Execute the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to install public key: %w (output: %s)", err, string(output))
	}

	return nil
}

// TestConnection tests an SSH connection
func (km *KeyManager) TestConnection(host, user, keyPath string, port int) error {
	// Read private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to remote server
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer client.Close()

	// Test with a simple command
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Run a simple test command
	if err := session.Run("echo 'SSH connection test successful'"); err != nil {
		return fmt.Errorf("failed to execute test command: %w", err)
	}

	return nil
}

// GetPublicKeyContent reads and returns the public key content
func (km *KeyManager) GetPublicKeyContent(keyPath string) (string, error) {
	pubKeyPath := keyPath + ".pub"
	data, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %w", err)
	}
	return string(data), nil
}

// DeployPublicKey deploys a public key to a remote server
func (km *KeyManager) DeployPublicKey(host string, port int, user, keyPath string) error {
	// First test if we can connect with the current key
	err := km.TestConnection(host, user, keyPath, port)
	if err == nil {
		// Connection already works, key might already be deployed
		return nil
	}

	// Try to install the key using ssh-copy-id equivalent
	return km.InstallPublicKey(host, user, keyPath, port)
}
