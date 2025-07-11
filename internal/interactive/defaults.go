package interactive

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os/user"
	"strings"
)

// Default values and word lists for tunnel creation
var (
	adjectives = []string{
		"swift", "clever", "bright", "calm", "bold", "quick", "sharp", "wise",
		"cool", "warm", "fast", "smooth", "strong", "light", "dark", "soft",
		"hard", "clear", "deep", "high", "low", "wide", "narrow", "thick",
		"thin", "heavy", "gentle", "rough", "smooth", "solid", "liquid",
		"crystal", "golden", "silver", "bronze", "steel", "iron", "copper",
		"digital", "cyber", "quantum", "neural", "atomic", "cosmic", "stellar",
		"lunar", "solar", "electric", "magnetic", "dynamic", "static", "active",
		"passive", "modern", "classic", "vintage", "future", "retro", "neo",
	}

	nouns = []string{
		"bridge", "tunnel", "gateway", "portal", "channel", "pathway", "conduit",
		"link", "connection", "network", "circuit", "wire", "cable", "fiber",
		"stream", "flow", "current", "wave", "signal", "pulse", "beam",
		"laser", "radar", "sonar", "scope", "lens", "mirror", "prism",
		"keyboard", "mouse", "screen", "display", "monitor", "terminal",
		"server", "client", "node", "host", "proxy", "router", "switch",
		"hub", "port", "socket", "adapter", "converter", "transformer",
		"engine", "motor", "generator", "reactor", "core", "processor",
		"memory", "storage", "cache", "buffer", "queue", "stack", "heap",
		"tree", "graph", "mesh", "grid", "matrix", "array", "vector",
	}
)

// GenerateRandomName generates a random tunnel name using two words separated by a hyphen
func GenerateRandomName() string {
	adjective := getRandomWord(adjectives)
	noun := getRandomWord(nouns)
	return fmt.Sprintf("%s-%s", adjective, noun)
}

// getRandomWord selects a random word from the given slice
func getRandomWord(words []string) string {
	if len(words) == 0 {
		return "tunnel"
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return words[0] // fallback to first word if random fails
	}

	return words[n.Int64()]
}

// GetDefaultUser returns the current system user as default
func GetDefaultUser() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "root" // fallback
}

// GetDefaultPort returns the default SSH port
func GetDefaultPort() string {
	return "22"
}

// ValidatePort checks if a port string is valid
func ValidatePort(portStr string) bool {
	if portStr == "" {
		return false
	}

	// Basic port validation - should be improved
	return !strings.Contains(portStr, " ") && len(portStr) > 0
}

// GetPortPlaceholder returns a helpful placeholder for port input
func GetPortPlaceholder() string {
	return "22 (default SSH port)"
}

// GetHostPlaceholder returns a helpful placeholder for host input
func GetHostPlaceholder() string {
	return "e.g., 192.168.1.100 or server.example.com"
}

// GetUserPlaceholder returns a helpful placeholder for username input
func GetUserPlaceholder() string {
	defaultUser := GetDefaultUser()
	return fmt.Sprintf("%s (current user)", defaultUser)
}

// GetNamePlaceholder returns a helpful placeholder for tunnel name
func GetNamePlaceholder() string {
	randomName := GenerateRandomName()
	return fmt.Sprintf("%s (random name)", randomName)
}
