package tokens

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// CountResult represents the result of token counting
type CountResult struct {
	FilePath    string         `json:"file_path"`
	IsDirectory bool           `json:"is_directory,omitempty"`
	FileCount   int            `json:"file_count,omitempty"`
	FileSize    int            `json:"file_size"`
	Characters  int            `json:"characters"`
	Words       int            `json:"words"`
	Lines       int            `json:"lines"`
	Methods     []MethodResult `json:"methods"`
	Costs       []CostEstimate `json:"costs,omitempty"`
}

// MethodResult represents token count for a specific method
type MethodResult struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Tokens      int    `json:"tokens"`
	IsExact     bool   `json:"is_exact"`
}

// CostEstimate represents cost estimation for a model
type CostEstimate struct {
	Model     string  `json:"model"`
	Tokens    int     `json:"tokens"`
	Cost      float64 `json:"cost"`
	RatePer1K float64 `json:"rate_per_1k"`
}

// CounterOptions configures the counter
type CounterOptions struct {
	CharsPerToken float64
	WordsPerToken float64
}

// Counter handles token counting
type Counter struct {
	charsPerToken float64
	wordsPerToken float64
	tokenizers    map[string]Tokenizer
}

// Tokenizer interface for different tokenization methods
type Tokenizer interface {
	CountTokens(text string) (int, error)
	Name() string        // Machine-readable identifier (e.g., "tiktoken_gpt_4")
	DisplayName() string // Human-readable name (e.g., "GPT (gpt-4)")
	IsExact() bool
}

// NewCounter creates a new token counter
func NewCounter(opts CounterOptions) *Counter {
	if opts.CharsPerToken == 0 {
		opts.CharsPerToken = 4.0
	}
	if opts.WordsPerToken == 0 {
		opts.WordsPerToken = 0.75
	}

	return &Counter{
		charsPerToken: opts.CharsPerToken,
		wordsPerToken: opts.WordsPerToken,
		tokenizers:    make(map[string]Tokenizer),
	}
}

// Count performs token counting using specified methods
func (c *Counter) Count(text string, model string, all bool) (*CountResult, error) {
	result := &CountResult{
		Characters: len(text),
		Words:      countWords(text),
		Lines:      countLines(text),
		Methods:    []MethodResult{},
	}

	// Initialize tokenizers if needed
	if err := c.initializeTokenizers(); err != nil {
		// Continue with approximations even if exact tokenizers fail
		// Silent failure - we'll use approximations
	}

	if all || model == "" {
		// Use all available methods
		result.Methods = c.countAllMethods(text)
	} else {
		// Use specific model
		methods, err := c.countSpecificModel(text, model)
		if err != nil {
			return nil, errors.Wrap(err, "counting tokens for model").WithField("model", model)
		}
		result.Methods = methods
	}

	return result, nil
}

// countAllMethods counts tokens using all available methods
func (c *Counter) countAllMethods(text string) []MethodResult {
	methods := []MethodResult{}

	// Try exact tokenizers first
	for _, tokenizer := range c.tokenizers {
		if count, err := tokenizer.CountTokens(text); err == nil {
			methods = append(methods, MethodResult{
				Name:        tokenizer.Name(),
				DisplayName: tokenizer.DisplayName(),
				Tokens:      count,
				IsExact:     tokenizer.IsExact(),
			})
		}
	}

	// Add approximation methods
	methods = append(methods, c.getApproximations(text)...)

	return methods
}

// countSpecificModel counts tokens for a specific model
func (c *Counter) countSpecificModel(text string, model string) ([]MethodResult, error) {
	methods := []MethodResult{}

	// Check if we have an exact tokenizer for this model
	if tokenizer, ok := c.tokenizers[model]; ok {
		count, err := tokenizer.CountTokens(text)
		if err != nil {
			return nil, err
		}
		methods = append(methods, MethodResult{
			Name:        tokenizer.Name(),
			DisplayName: tokenizer.DisplayName(),
			Tokens:      count,
			IsExact:     tokenizer.IsExact(),
		})
	} else {
		// Fall back to approximations
		methods = append(methods, c.getApproximations(text)...)
	}

	return methods, nil
}

// getApproximations returns approximation-based token counts
func (c *Counter) getApproximations(text string) []MethodResult {
	chars := len(text)
	words := countWords(text)

	// Format multiplier for word-based calculation
	multiplier := 1.0 / c.wordsPerToken
	multiplierStr := fmt.Sprintf("%.0f", multiplier*100) // e.g., "133" for 1.33

	return []MethodResult{
		{
			Name:        fmt.Sprintf("character_based_div%.0f", c.charsPerToken),
			DisplayName: fmt.Sprintf("Character-based (÷%.1f)", c.charsPerToken),
			Tokens:      int(float64(chars) / c.charsPerToken),
			IsExact:     false,
		},
		{
			Name:        fmt.Sprintf("word_based_mul%s", multiplierStr),
			DisplayName: fmt.Sprintf("Word-based (×%.2f)", multiplier),
			Tokens:      int(float64(words) / c.wordsPerToken),
			IsExact:     false,
		},
		{
			Name:        "whitespace_split",
			DisplayName: "Whitespace split",
			Tokens:      words,
			IsExact:     false,
		},
	}
}

// initializeTokenizers sets up available tokenizers
func (c *Counter) initializeTokenizers() error {
	// Try to initialize tiktoken for GPT models
	if tokenizer, err := NewTiktokenTokenizer("gpt-4"); err == nil {
		c.tokenizers["gpt-4"] = tokenizer
	}

	if tokenizer, err := NewTiktokenTokenizer("gpt-3.5-turbo"); err == nil {
		c.tokenizers["gpt-3.5-turbo"] = tokenizer
	}

	// Add Claude approximation (since we don't have offline tokenizer)
	c.tokenizers["claude-3"] = NewClaudeApproximator()

	return nil
}

// countWords counts words in text
func countWords(text string) int {
	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		words++
	}

	return words
}

// countLines counts lines in text
func countLines(text string) int {
	if len(text) == 0 {
		return 0
	}

	lines := strings.Count(text, "\n")
	// Add 1 if the last character is not a newline
	if text[len(text)-1] != '\n' {
		lines++
	}

	return lines
}
