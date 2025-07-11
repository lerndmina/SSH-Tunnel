package interactive

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lerndmina/SSH-Tunnel/internal/config"
	"github.com/lerndmina/SSH-Tunnel/internal/ssh"
	"github.com/lerndmina/SSH-Tunnel/internal/tunnel"
	cryptossh "golang.org/x/crypto/ssh"
)

// SimpleTUI provides a simple command-line interface for tunnel management
type SimpleTUI struct {
	keyManager  *ssh.KeyManager
	tunnelMgr   *tunnel.Manager
	configMgr   *config.Manager
	scanner     *bufio.Scanner
}

// Colors for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func colorize(text, color string) string {
	return color + text + colorReset
}

// NewSimpleTUI creates a new simple TUI instance
func NewSimpleTUI() (*SimpleTUI, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	
	configPath := filepath.Join(homeDir, ".ssh-tunnel-manager")
	configMgr, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %v", err)
	}

	return &SimpleTUI{
		keyManager: ssh.NewKeyManager(),
		tunnelMgr:  tunnel.NewManager(),
		configMgr:  configMgr,
		scanner:    bufio.NewScanner(os.Stdin),
	}, nil
}

// Run starts the interactive tunnel creation process
func (tui *SimpleTUI) Run() error {
	fmt.Println(colorize("=== SSH Tunnel Manager ===", colorCyan))
	fmt.Println()

	for {
		fmt.Println(colorize("Main Menu:", colorBlue))
		fmt.Println("1) Create new tunnel")
		fmt.Println("2) List tunnels")
		fmt.Println("3) Start tunnel")
		fmt.Println("4) Stop tunnel")
		fmt.Println("5) Delete tunnel")
		fmt.Println("6) Exit")
		fmt.Println()

		choice, err := tui.promptString("Select an option (1-6)", "", true)
		if err != nil {
			return err
		}

		switch choice {
		case "1":
			if err := tui.createNewTunnel(); err != nil {
				fmt.Printf("%s: %v\n", colorize("Error", colorRed), err)
				tui.promptContinue()
			}
		case "2":
			tui.listTunnels()
		case "3":
			if err := tui.startTunnel(); err != nil {
				fmt.Printf("%s: %v\n", colorize("Error", colorRed), err)
				tui.promptContinue()
			}
		case "4":
			if err := tui.stopTunnel(); err != nil {
				fmt.Printf("%s: %v\n", colorize("Error", colorRed), err)
				tui.promptContinue()
			}
		case "5":
			if err := tui.deleteTunnel(); err != nil {
				fmt.Printf("%s: %v\n", colorize("Error", colorRed), err)
				tui.promptContinue()
			}
		case "6":
			fmt.Println(colorize("Goodbye!", colorGreen))
			return nil
		default:
			fmt.Println(colorize("Invalid choice. Please try again.", colorRed))
		}
		fmt.Println()
	}
}

func (tui *SimpleTUI) createNewTunnel() error {
	fmt.Println(colorize("=== Create New Tunnel ===", colorCyan))
	fmt.Println()

	// Get tunnel configuration
	cfg, err := tui.promptForTunnelConfig()
	if err != nil {
		return err
	}

	// Setup SSH key
	if err := tui.setupSSHKey(cfg); err != nil {
		return err
	}

	// Setup natted server connection
	if err := tui.setupNattedServerConnection(cfg); err != nil {
		return err
	}

	// Save configuration
	if err := tui.configMgr.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save tunnel configuration: %v", err)
	}

	fmt.Println(colorize("Tunnel created successfully!", colorGreen))
	fmt.Printf("Tunnel '%s' is ready to use.\n", cfg.TunnelName)
	fmt.Println()

	startNow, err := tui.promptYesNo("Would you like to start the tunnel now?", true)
	if err != nil {
		return err
	}

	if startNow {
		if err := tui.tunnelMgr.Start(cfg.TunnelName); err != nil {
			return fmt.Errorf("failed to start tunnel: %v", err)
		}
		fmt.Println(colorize("Tunnel started successfully!", colorGreen))
	}

	return nil
}

func (tui *SimpleTUI) promptForTunnelConfig() (*config.Config, error) {
	cfg := &config.Config{}
	var err error

	// Generate random name as default
	defaultName := tui.generateRandomTunnelName()
	cfg.TunnelName, err = tui.promptString("Tunnel Name", defaultName, true)
	if err != nil {
		return nil, err
	}

	cfg.CloudServer.IP, err = tui.promptString("Cloud Server IP", "", true)
	if err != nil {
		return nil, err
	}

	// Cloud SSH Port with default
	cloudSSHPortStr, err := tui.promptString("Cloud SSH Port", "22", true)
	if err != nil {
		return nil, err
	}
	cfg.CloudServer.Port, err = strconv.Atoi(cloudSSHPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cloud SSH port: %v", err)
	}

	cfg.CloudServer.User, err = tui.promptString("Cloud User", "root", true)
	if err != nil {
		return nil, err
	}

	// Reverse Port with sensible default
	reversePortStr, err := tui.promptString("Reverse Port", "2222", true)
	if err != nil {
		return nil, err
	}
	cfg.LocalServer.ReversePort, err = strconv.Atoi(reversePortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid reverse port: %v", err)
	}

	// Get current user as default for NattedUser
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = os.Getenv("USERNAME") // Windows
	}
	if currentUser == "" {
		currentUser = "user"
	}

	cfg.LocalServer.User, err = tui.promptString("Natted User", currentUser, true)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (tui *SimpleTUI) setupSSHKey(cfg *config.Config) error {
	fmt.Println()
	fmt.Println(colorize("SSH Private Key Setup", colorYellow))
	fmt.Println("Choose an option for SSH authentication:")
	fmt.Println("1) Paste private key content directly")
	fmt.Println("2) Provide path to existing private key file")
	fmt.Println("3) Generate new SSH key pair")
	fmt.Println("4) Use existing key at ~/.ssh/cloud_server_key")
	fmt.Println()

	// Check if the default key already exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	defaultKeyPath := filepath.Join(homeDir, ".ssh", "cloud_server_key")
	
	if _, err := os.Stat(defaultKeyPath); err == nil {
		fmt.Println(colorize("Found existing key at "+defaultKeyPath, colorGreen))
	} else {
		fmt.Println(colorize("No existing key found at "+defaultKeyPath, colorYellow))
	}
	fmt.Println()

	var keyChoice string
	for {
		keyChoice, err = tui.promptString("Enter choice (1-4)", "", true)
		if err != nil {
			return err
		}
		if keyChoice == "1" || keyChoice == "2" || keyChoice == "3" || keyChoice == "4" {
			break
		}
		fmt.Println(colorize("Invalid choice. Please enter 1, 2, 3, or 4.", colorRed))
	}

	// Create SSH directory if it doesn't exist
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %v", err)
	}

	privateKeyPath := filepath.Join(sshDir, "cloud_server_key")

	switch keyChoice {
	case "1":
		fmt.Println("Paste your private key content (press Ctrl+D when finished):")
		content, err := tui.readMultilineInput()
		if err != nil {
			return fmt.Errorf("failed to read private key content: %v", err)
		}
		if err := os.WriteFile(privateKeyPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("failed to write private key: %v", err)
		}
		fmt.Println(colorize("Private key saved successfully", colorGreen))

	case "2":
		var existingKeyPath string
		for {
			existingKeyPath, err = tui.promptString("Enter path to private key file", "", true)
			if err != nil {
				return err
			}
			if _, err := os.Stat(existingKeyPath); err == nil {
				// Check if it's the same file
				absExisting, _ := filepath.Abs(existingKeyPath)
				absTarget, _ := filepath.Abs(privateKeyPath)
				if absExisting == absTarget {
					fmt.Println(colorize("Using existing key at "+existingKeyPath, colorGreen))
				} else {
					content, err := os.ReadFile(existingKeyPath)
					if err != nil {
						return fmt.Errorf("failed to read key file: %v", err)
					}
					if err := os.WriteFile(privateKeyPath, content, 0600); err != nil {
						return fmt.Errorf("failed to copy key: %v", err)
					}
					fmt.Println(colorize("Key copied to "+privateKeyPath, colorGreen))
				}
				break
			} else {
				fmt.Println(colorize("File not found. Please try again.", colorRed))
			}
		}

	case "3":
		fmt.Println(colorize("Generating new SSH key pair...", colorYellow))
		if err := tui.keyManager.GenerateKeyPair("ed25519", privateKeyPath); err != nil {
			return fmt.Errorf("failed to generate key pair: %v", err)
		}
		fmt.Println(colorize("New SSH key pair generated!", colorGreen))
		fmt.Println(colorize("IMPORTANT: Copy the following public key to your cloud server's ~/.ssh/authorized_keys:", colorYellow))
		fmt.Println()
		
		pubKeyContent, err := os.ReadFile(privateKeyPath + ".pub")
		if err != nil {
			return fmt.Errorf("failed to read public key: %v", err)
		}
		fmt.Print(string(pubKeyContent))
		fmt.Println()
		
		_, err = tui.promptString("Press Enter after you've added the public key to your cloud server...", "", false)
		if err != nil {
			return err
		}

	case "4":
		if _, err := os.Stat(defaultKeyPath); err == nil {
			fmt.Println(colorize("Using existing key at "+defaultKeyPath, colorGreen))
			if err := os.Chmod(privateKeyPath, 0600); err != nil {
				return fmt.Errorf("failed to set key permissions: %v", err)
			}
		} else {
			return fmt.Errorf("key not found at %s. Please choose another option", defaultKeyPath)
		}
	}

	// Test SSH connection
	fmt.Println(colorize("Testing SSH connection to cloud server...", colorYellow))
	fmt.Println("Testing connection (you may need to accept the host key)...")
	
	if err := tui.keyManager.TestConnection(cfg.CloudServer.IP, cfg.CloudServer.User, privateKeyPath, cfg.CloudServer.Port); err != nil {
		fmt.Println(colorize("SSH connection failed. Please check your credentials and try again.", colorRed))
		return fmt.Errorf("SSH connection test failed: %v", err)
	}
	fmt.Println(colorize("SSH connection successful!", colorGreen))

	return nil
}

func (tui *SimpleTUI) setupNattedServerConnection(cfg *config.Config) error {
	fmt.Println(colorize("Setting up connection from cloud server to NAT'd server...", colorYellow))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	nattedKeyPath := filepath.Join(sshDir, fmt.Sprintf("natted_server_key_%s", cfg.TunnelName))

	// Generate a separate key pair for connecting FROM cloud server TO NAT'd server
	fmt.Println("Generating SSH key pair for cloud server to connect to NAT'd server...")
	if err := tui.keyManager.GenerateKeyPair("ed25519", nattedKeyPath); err != nil {
		return fmt.Errorf("failed to generate natted server key pair: %v", err)
	}

	// Add the public key to this server's authorized_keys
	fmt.Println("Adding public key to this server's authorized_keys...")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %v", err)
	}

	authorizedKeysPath := filepath.Join(sshDir, "authorized_keys")
	pubKeyContent, err := os.ReadFile(nattedKeyPath + ".pub")
	if err != nil {
		return fmt.Errorf("failed to read public key: %v", err)
	}

	// Check if key already exists in authorized_keys
	if _, err := os.Stat(authorizedKeysPath); err == nil {
		existingKeys, err := os.ReadFile(authorizedKeysPath)
		if err != nil {
			return fmt.Errorf("failed to read authorized_keys: %v", err)
		}
		if strings.Contains(string(existingKeys), strings.TrimSpace(string(pubKeyContent))) {
			fmt.Println(colorize("Public key already exists in authorized_keys", colorYellow))
		} else {
			if err := tui.appendToAuthorizedKeys(authorizedKeysPath, pubKeyContent); err != nil {
				return err
			}
			fmt.Println(colorize("Public key added to authorized_keys", colorGreen))
		}
	} else {
		if err := tui.appendToAuthorizedKeys(authorizedKeysPath, pubKeyContent); err != nil {
			return err
		}
		fmt.Println(colorize("Public key added to authorized_keys", colorGreen))
	}

	// Deploy the natted server's private key to cloud server so it can connect back
	fmt.Println("Deploying natted server private key to cloud server...")
	cloudKeyPath := filepath.Join(homeDir, ".ssh", "cloud_server_key")
	if err := tui.deployNattedKeyToCloud(cfg.CloudServer.IP, cfg.CloudServer.Port, cfg.CloudServer.User, cloudKeyPath, nattedKeyPath); err != nil {
		return fmt.Errorf("failed to deploy natted key to cloud server: %v", err)
	}

	// Create connection script on cloud server
	fmt.Println("Creating connection script on cloud server...")
	if err := tui.createConnectionScript(cfg.CloudServer.IP, cfg.CloudServer.Port, cfg.CloudServer.User, cloudKeyPath, nattedKeyPath, cfg); err != nil {
		return fmt.Errorf("failed to create connection script: %v", err)
	}

	// Set the SSH config paths
	cfg.SSH.PrivateKeyPath = filepath.Join(homeDir, ".ssh", "cloud_server_key")
	cfg.SSH.NattedKeyPath = nattedKeyPath

	return nil
}

func (tui *SimpleTUI) appendToAuthorizedKeys(authorizedKeysPath string, pubKeyContent []byte) error {
	file, err := os.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(pubKeyContent); err != nil {
		return fmt.Errorf("failed to write to authorized_keys: %v", err)
	}

	// Ensure it ends with a newline
	if !strings.HasSuffix(string(pubKeyContent), "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline to authorized_keys: %v", err)
		}
	}

	return nil
}

func (tui *SimpleTUI) readMultilineInput() (string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return "", err
	}
	
	return strings.Join(lines, "\n"), nil
}

func (tui *SimpleTUI) listTunnels() {
	fmt.Println(colorize("=== Tunnel List ===", colorCyan))
	
	tunnelNames := tui.configMgr.ListConfigs()
	if len(tunnelNames) == 0 {
		fmt.Println("No tunnels found.")
		tui.promptContinue()
		return
	}

	for i, tunnelName := range tunnelNames {
		cfg, err := tui.configMgr.GetConfig(tunnelName)
		if err != nil {
			fmt.Printf("%s: Failed to load tunnel '%s': %v\n", colorize("Error", colorRed), tunnelName, err)
			continue
		}

		status, err := tui.tunnelMgr.GetStatus(tunnelName)
		statusStr := "Stopped"
		if err == nil && status != nil && status.Status == tunnel.StatusRunning {
			statusStr = colorize("Running", colorGreen)
		} else {
			statusStr = colorize("Stopped", colorRed)
		}

		fmt.Printf("%d. %s (%s)\n", i+1, cfg.TunnelName, statusStr)
		fmt.Printf("   Cloud: %s@%s:%d -> Reverse: %d\n", 
			cfg.CloudServer.User, cfg.CloudServer.IP, cfg.CloudServer.Port, cfg.LocalServer.ReversePort)
		fmt.Printf("   Natted User: %s\n", cfg.LocalServer.User)
		fmt.Println()
	}

	tui.promptContinue()
}

func (tui *SimpleTUI) startTunnel() error {
	tunnelNames := tui.configMgr.ListConfigs()
	if len(tunnelNames) == 0 {
		fmt.Println("No tunnels found.")
		tui.promptContinue()
		return nil
	}

	fmt.Println(colorize("=== Start Tunnel ===", colorCyan))
	for i, tunnelName := range tunnelNames {
		status, err := tui.tunnelMgr.GetStatus(tunnelName)
		statusStr := "Stopped"
		if err == nil && status != nil && status.Status == tunnel.StatusRunning {
			statusStr = colorize("Running", colorGreen)
		} else {
			statusStr = colorize("Stopped", colorRed)
		}
		fmt.Printf("%d. %s (%s)\n", i+1, tunnelName, statusStr)
	}

	choice, err := tui.promptString("Select tunnel to start (number)", "", true)
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(tunnelNames) {
		return fmt.Errorf("invalid selection")
	}

	selectedTunnel := tunnelNames[index-1]
	
	status, err := tui.tunnelMgr.GetStatus(selectedTunnel)
	if err == nil && status != nil && status.Status == tunnel.StatusRunning {
		fmt.Printf("Tunnel '%s' is already running.\n", selectedTunnel)
		tui.promptContinue()
		return nil
	}

	if err := tui.tunnelMgr.Start(selectedTunnel); err != nil {
		return fmt.Errorf("failed to start tunnel: %v", err)
	}

	fmt.Printf("%s started successfully!\n", colorize("Tunnel '"+selectedTunnel+"'", colorGreen))
	tui.promptContinue()
	return nil
}

func (tui *SimpleTUI) stopTunnel() error {
	tunnelNames := tui.configMgr.ListConfigs()
	if len(tunnelNames) == 0 {
		fmt.Println("No tunnels found.")
		tui.promptContinue()
		return nil
	}

	fmt.Println(colorize("=== Stop Tunnel ===", colorCyan))
	for i, tunnelName := range tunnelNames {
		status, err := tui.tunnelMgr.GetStatus(tunnelName)
		statusStr := "Stopped"
		if err == nil && status != nil && status.Status == tunnel.StatusRunning {
			statusStr = colorize("Running", colorGreen)
		} else {
			statusStr = colorize("Stopped", colorRed)
		}
		fmt.Printf("%d. %s (%s)\n", i+1, tunnelName, statusStr)
	}

	choice, err := tui.promptString("Select tunnel to stop (number)", "", true)
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(tunnelNames) {
		return fmt.Errorf("invalid selection")
	}

	selectedTunnel := tunnelNames[index-1]
	
	status, err := tui.tunnelMgr.GetStatus(selectedTunnel)
	if err != nil || status == nil || status.Status != tunnel.StatusRunning {
		fmt.Printf("Tunnel '%s' is not running.\n", selectedTunnel)
		tui.promptContinue()
		return nil
	}

	if err := tui.tunnelMgr.Stop(selectedTunnel); err != nil {
		return fmt.Errorf("failed to stop tunnel: %v", err)
	}

	fmt.Printf("%s stopped successfully!\n", colorize("Tunnel '"+selectedTunnel+"'", colorGreen))
	tui.promptContinue()
	return nil
}

func (tui *SimpleTUI) deleteTunnel() error {
	tunnelNames := tui.configMgr.ListConfigs()
	if len(tunnelNames) == 0 {
		fmt.Println("No tunnels found.")
		tui.promptContinue()
		return nil
	}

	fmt.Println(colorize("=== Delete Tunnel ===", colorCyan))
	for i, tunnelName := range tunnelNames {
		status, err := tui.tunnelMgr.GetStatus(tunnelName)
		statusStr := "Stopped"
		if err == nil && status != nil && status.Status == tunnel.StatusRunning {
			statusStr = colorize("Running", colorGreen)
		} else {
			statusStr = colorize("Stopped", colorRed)
		}
		fmt.Printf("%d. %s (%s)\n", i+1, tunnelName, statusStr)
	}

	choice, err := tui.promptString("Select tunnel to delete (number)", "", true)
	if err != nil {
		return err
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(tunnelNames) {
		return fmt.Errorf("invalid selection")
	}

	selectedTunnel := tunnelNames[index-1]

	// Confirm deletion
	confirmed, err := tui.promptYesNo(fmt.Sprintf("Are you sure you want to delete tunnel '%s'?", selectedTunnel), false)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Deletion cancelled.")
		tui.promptContinue()
		return nil
	}

	// Stop tunnel if running
	status, err := tui.tunnelMgr.GetStatus(selectedTunnel)
	if err == nil && status != nil && status.Status == tunnel.StatusRunning {
		if err := tui.tunnelMgr.Stop(selectedTunnel); err != nil {
			return fmt.Errorf("failed to stop tunnel before deletion: %v", err)
		}
	}

	// Delete tunnel configuration
	if err := tui.configMgr.DeleteConfig(selectedTunnel); err != nil {
		return fmt.Errorf("failed to delete tunnel: %v", err)
	}

	fmt.Printf("%s deleted successfully!\n", colorize("Tunnel '"+selectedTunnel+"'", colorGreen))
	tui.promptContinue()
	return nil
}

func (tui *SimpleTUI) promptString(prompt, defaultValue string, required bool) (string, error) {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		if !tui.scanner.Scan() {
			return "", fmt.Errorf("failed to read input")
		}

		input := strings.TrimSpace(tui.scanner.Text())
		if input == "" && defaultValue != "" {
			return defaultValue, nil
		}
		if input != "" || !required {
			return input, nil
		}

		fmt.Println(colorize("This field is required. Please enter a value.", colorRed))
	}
}

func (tui *SimpleTUI) promptYesNo(prompt string, defaultValue bool) (bool, error) {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		response, err := tui.promptString(prompt+" (y/n)", defaultStr, true)
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		switch response {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println(colorize("Please enter 'y' or 'n'.", colorRed))
		}
	}
}

func (tui *SimpleTUI) promptContinue() {
	fmt.Print("Press Enter to continue...")
	tui.scanner.Scan()
}

func (tui *SimpleTUI) generateRandomTunnelName() string {
	adjectives := []string{"rapid", "secure", "swift", "stable", "direct", "fast", "safe", "quick", "solid", "smooth"}
	nouns := []string{"tunnel", "bridge", "link", "pipe", "channel", "path", "route", "gateway", "portal", "conduit"}
	
	// Simple random selection based on time
	adjIndex := len(adjectives) % 7  // Use a simple hash
	nounIndex := len(nouns) % 5
	
	return fmt.Sprintf("%s-%s", adjectives[adjIndex], nouns[nounIndex])
}

func (tui *SimpleTUI) deployNattedKeyToCloud(cloudHost string, cloudPort int, cloudUser, cloudKeyPath, nattedKeyPath string) error {
	// Read the cloud server private key for authentication
	cloudKeyData, err := os.ReadFile(cloudKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read cloud server key: %w", err)
	}

	cloudSigner, err := cryptossh.ParsePrivateKey(cloudKeyData)
	if err != nil {
		return fmt.Errorf("failed to parse cloud server key: %w", err)
	}

	// Read the natted server private key to deploy
	nattedKeyData, err := os.ReadFile(nattedKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read natted server key: %w", err)
	}

	// Create SSH client config using cloud server key
	config := &cryptossh.ClientConfig{
		User: cloudUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.PublicKeys(cloudSigner),
		},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to cloud server
	address := net.JoinHostPort(cloudHost, fmt.Sprintf("%d", cloudPort))
	client, err := cryptossh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to cloud server: %w", err)
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create the .ssh directory and deploy the natted server key
	nattedKeyB64 := base64.StdEncoding.EncodeToString(nattedKeyData)
	keyFileName := filepath.Base(nattedKeyPath)
	
	cmd := fmt.Sprintf(`
		mkdir -p ~/.ssh &&
		chmod 700 ~/.ssh &&
		echo '%s' | base64 -d > ~/.ssh/%s &&
		chmod 600 ~/.ssh/%s &&
		echo "Natted server key deployed successfully"
	`, nattedKeyB64, keyFileName, keyFileName)

	// Execute the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to deploy natted key: %w (output: %s)", err, string(output))
	}

	fmt.Println(colorize("Natted server private key deployed to cloud server", colorGreen))
	return nil
}

func (tui *SimpleTUI) createConnectionScript(cloudHost string, cloudPort int, cloudUser, cloudKeyPath, nattedKeyPath string, cfg *config.Config) error {
	// Read the cloud server private key for authentication
	cloudKeyData, err := os.ReadFile(cloudKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read cloud server key: %w", err)
	}

	cloudSigner, err := cryptossh.ParsePrivateKey(cloudKeyData)
	if err != nil {
		return fmt.Errorf("failed to parse cloud server key: %w", err)
	}

	// Create SSH client config using cloud server key
	config := &cryptossh.ClientConfig{
		User: cloudUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.PublicKeys(cloudSigner),
		},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to cloud server
	address := net.JoinHostPort(cloudHost, fmt.Sprintf("%d", cloudPort))
	client, err := cryptossh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to cloud server: %w", err)
	}
	defer client.Close()

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Create the connection script
	nattedKeyFileName := filepath.Base(nattedKeyPath)
	scriptContent := fmt.Sprintf(`#!/bin/bash
# Connection script for reverse SSH tunnel: %s
# This script connects from cloud server to NAT'd server

NATTED_HOST="localhost"  # Connection will be via reverse tunnel
NATTED_PORT="22"         # Standard SSH port on local machine
NATTED_USER="%s"
NATTED_KEY="~/.ssh/%s"
REVERSE_PORT="%d"

# Function to establish connection via reverse tunnel
connect_via_reverse_tunnel() {
    echo "Connecting to NAT'd server via reverse tunnel on port ${REVERSE_PORT}..."
    ssh -i "$NATTED_KEY" \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -p ${REVERSE_PORT} \
        ${NATTED_USER}@localhost
}

# Function to test connection via reverse tunnel
test_reverse_connection() {
    echo "Testing connection to NAT'd server via reverse tunnel..."
    ssh -i "$NATTED_KEY" \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -o ConnectTimeout=10 \
        -p ${REVERSE_PORT} \
        ${NATTED_USER}@localhost \
        "echo 'Reverse tunnel connection successful!'"
}

case "$1" in
    test)
        test_reverse_connection
        ;;
    connect)
        connect_via_reverse_tunnel
        ;;
    *)
        echo "Usage: $0 {test|connect}"
        echo "  test    - Test connection to NAT'd server via reverse tunnel"
        echo "  connect - Connect to NAT'd server via reverse tunnel"
        echo ""
        echo "Note: This script assumes the reverse tunnel is already established"
        echo "Reverse tunnel port: ${REVERSE_PORT}"
        exit 1
        ;;
esac
`, cfg.TunnelName, cfg.LocalServer.User, nattedKeyFileName, cfg.LocalServer.ReversePort)

	scriptPath := fmt.Sprintf("connect_%s.sh", cfg.TunnelName)
	
	cmd := fmt.Sprintf(`
		cat > %s << 'SCRIPT_EOF'
%s
SCRIPT_EOF
		chmod +x %s &&
		echo "Connection script created successfully at %s"
	`, scriptPath, scriptContent, scriptPath, scriptPath)

	// Execute the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to create connection script: %w (output: %s)", err, string(output))
	}

	fmt.Println(colorize(fmt.Sprintf("Connection script created at %s on cloud server", scriptPath), colorGreen))
	fmt.Println(colorize(fmt.Sprintf("To test connection: ./%s test", scriptPath), colorCyan))
	fmt.Println(colorize(fmt.Sprintf("To start tunnel: ./%s connect", scriptPath), colorCyan))
	return nil
}
