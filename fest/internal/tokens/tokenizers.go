package tokens

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// TiktokenTokenizer implements exact tokenization for OpenAI models
type TiktokenTokenizer struct {
	model    string
	encoding *tiktoken.Tiktoken
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer
func NewTiktokenTokenizer(model string) (*TiktokenTokenizer, error) {
	// Map model names to encodings
	encodingName := getEncodingForModel(model)

	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		// Try to get encoding for the model directly
		encoding, err = tiktoken.EncodingForModel(model)
		if err != nil {
			return nil, errors.Wrap(err, "getting encoding for model").WithField("model", model)
		}
	}

	return &TiktokenTokenizer{
		model:    model,
		encoding: encoding,
	}, nil
}

// CountTokens counts tokens using tiktoken
func (t *TiktokenTokenizer) CountTokens(text string) (int, error) {
	tokens := t.encoding.Encode(text, nil, nil)
	return len(tokens), nil
}

// Name returns the machine-readable tokenizer identifier
func (t *TiktokenTokenizer) Name() string {
	// Convert model name to snake_case identifier
	// e.g., "gpt-4" -> "tiktoken_gpt_4"
	//       "gpt-3.5-turbo" -> "tiktoken_gpt_3_5_turbo"
	modelName := strings.ReplaceAll(t.model, "-", "_")
	modelName = strings.ReplaceAll(modelName, ".", "_")
	return fmt.Sprintf("tiktoken_%s", modelName)
}

// DisplayName returns the human-readable tokenizer name
func (t *TiktokenTokenizer) DisplayName() string {
	return fmt.Sprintf("GPT (%s)", t.model)
}

// IsExact returns true for tiktoken tokenizers
func (t *TiktokenTokenizer) IsExact() bool {
	return true
}

// getEncodingForModel maps model names to encoding types
func getEncodingForModel(model string) string {
	model = strings.ToLower(model)

	// GPT-4 and GPT-3.5-turbo use cl100k_base encoding
	if strings.Contains(model, "gpt-4") || strings.Contains(model, "gpt-3.5") {
		return "cl100k_base"
	}

	// Older models
	if strings.Contains(model, "davinci") || strings.Contains(model, "curie") {
		return "p50k_base"
	}

	// Default to cl100k_base for newer models
	return "cl100k_base"
}

// ClaudeApproximator provides approximation for Claude models
type ClaudeApproximator struct{}

// NewClaudeApproximator creates a new Claude approximator
func NewClaudeApproximator() *ClaudeApproximator {
	return &ClaudeApproximator{}
}

// CountTokens approximates token count for Claude
func (c *ClaudeApproximator) CountTokens(text string) (int, error) {
	// Claude generally uses similar tokenization to GPT models
	// We'll use a slightly adjusted character-based approximation
	// Claude tends to use slightly more tokens than GPT for the same text
	chars := len(text)
	tokens := int(float64(chars) / 3.8) // Slightly more tokens than the 4.0 ratio

	return tokens, nil
}

// Name returns the machine-readable tokenizer identifier
func (c *ClaudeApproximator) Name() string {
	return "claude_3_approx"
}

// DisplayName returns the human-readable tokenizer name
func (c *ClaudeApproximator) DisplayName() string {
	return "Claude-3 (approx)"
}

// IsExact returns false for approximations
func (c *ClaudeApproximator) IsExact() bool {
	return false
}

// SimpleTokenizer provides basic tokenization
type SimpleTokenizer struct {
	name        string
	displayName string
	method      string
}

// NewSimpleTokenizer creates a basic tokenizer
func NewSimpleTokenizer(name, displayName, method string) *SimpleTokenizer {
	return &SimpleTokenizer{
		name:        name,
		displayName: displayName,
		method:      method,
	}
}

// CountTokens counts tokens using simple methods
func (s *SimpleTokenizer) CountTokens(text string) (int, error) {
	switch s.method {
	case "whitespace":
		return len(strings.Fields(text)), nil
	case "chars4":
		return len(text) / 4, nil
	case "words1.33":
		words := len(strings.Fields(text))
		return int(float64(words) * 1.33), nil
	default:
		return 0, errors.Validation("unknown tokenization method").WithField("method", s.method)
	}
}

// Name returns the machine-readable tokenizer identifier
func (s *SimpleTokenizer) Name() string {
	return s.name
}

// DisplayName returns the human-readable tokenizer name
func (s *SimpleTokenizer) DisplayName() string {
	return s.displayName
}

// IsExact returns false for simple tokenizers
func (s *SimpleTokenizer) IsExact() bool {
	return false
}
