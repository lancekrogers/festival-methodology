// Package id provides unique identifier generation for festival elements.
package id

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	// IDPrefix is the prefix for all festival IDs
	IDPrefix = "FEST"
	// IDLength is the length of the hash portion
	IDLength = 6
)

var (
	// IDPattern matches a valid festival ID
	IDPattern = regexp.MustCompile(`^FEST-[a-z0-9]{6}$`)
)

// Generator creates unique festival IDs
type Generator struct {
	mu   sync.RWMutex
	seen map[string]bool
}

// NewGenerator creates a new ID generator
func NewGenerator() *Generator {
	return &Generator{
		seen: make(map[string]bool),
	}
}

// Generate creates a unique ID from a path and timestamp
func (g *Generator) Generate(path string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Hash input: relative path + creation timestamp
	timestamp := time.Now().UnixNano()
	input := fmt.Sprintf("%s:%d", path, timestamp)

	for attempt := 0; attempt < 100; attempt++ {
		hash := sha256.Sum256([]byte(input))
		hexStr := hex.EncodeToString(hash[:])

		// Take first IDLength chars, lowercase
		id := fmt.Sprintf("%s-%s", IDPrefix, strings.ToLower(hexStr[:IDLength]))

		// Check for collision
		if !g.seen[id] {
			g.seen[id] = true
			return id
		}

		// Add attempt counter for next iteration
		input = fmt.Sprintf("%s:%d:%d", path, timestamp, attempt+1)
	}

	// Fallback with more entropy
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d", path, time.Now().UnixNano(), time.Now().Nanosecond())))
	return fmt.Sprintf("%s-%s", IDPrefix, strings.ToLower(hex.EncodeToString(hash[:])[:IDLength]))
}

// GenerateFromContent creates an ID based on content hash (deterministic)
func (g *Generator) GenerateFromContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	hexStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s-%s", IDPrefix, strings.ToLower(hexStr[:IDLength]))
}

// Register marks an ID as already used
func (g *Generator) Register(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.seen[id] = true
}

// IsRegistered checks if an ID is already used
func (g *Generator) IsRegistered(id string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.seen[id]
}

// Validate checks if a string is a valid festival ID
func Validate(id string) bool {
	return IDPattern.MatchString(id)
}

// Parse extracts the hash portion from a festival ID
func Parse(id string) (string, error) {
	if !Validate(id) {
		return "", fmt.Errorf("invalid festival ID format: %s", id)
	}
	return strings.TrimPrefix(id, IDPrefix+"-"), nil
}

// Format creates a properly formatted ID from a hash
func Format(hash string) string {
	// Ensure lowercase and correct length
	h := strings.ToLower(hash)
	if len(h) > IDLength {
		h = h[:IDLength]
	}
	return fmt.Sprintf("%s-%s", IDPrefix, h)
}

// ExtractFromMessage extracts festival IDs from a commit message
func ExtractFromMessage(message string) []string {
	pattern := regexp.MustCompile(`\[FEST-[a-z0-9]{6}\]`)
	matches := pattern.FindAllString(message, -1)

	var ids []string
	for _, match := range matches {
		// Remove brackets
		id := strings.TrimPrefix(match, "[")
		id = strings.TrimSuffix(id, "]")
		ids = append(ids, id)
	}
	return ids
}

// FormatForCommit formats an ID for inclusion in a commit message
func FormatForCommit(id string) string {
	return fmt.Sprintf("[%s]", id)
}
