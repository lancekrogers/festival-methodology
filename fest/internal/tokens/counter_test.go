package tokens

import (
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Empty", "", 0},
		{"Single word", "hello", 1},
		{"Multiple words", "hello world", 2},
		{"With punctuation", "Hello, world!", 2},
		{"Multiple spaces", "hello   world", 2},
		{"With newlines", "hello\nworld", 2},
		{"Mixed whitespace", "hello\t\nworld  test", 3},
		{"Code sample", "func main() { fmt.Println(\"hello\") }", 5},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"Empty", "", 0},
		{"Single line", "hello", 1},
		{"Two lines", "hello\nworld", 2},
		{"Trailing newline", "hello\n", 1},
		{"Multiple newlines", "hello\n\nworld", 3},
		{"Windows line endings", "hello\r\nworld", 2},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countLines(tt.text)
			if result != tt.expected {
				t.Errorf("countLines(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestCounter_Count(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."
	
	counter := NewCounter(CounterOptions{
		CharsPerToken: 4.0,
		WordsPerToken: 0.75,
	})
	
	result, err := counter.Count(text, "", true)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	
	// Check basic stats
	if result.Characters != len(text) {
		t.Errorf("Characters = %d, want %d", result.Characters, len(text))
	}
	
	if result.Words != 9 {
		t.Errorf("Words = %d, want 9", result.Words)
	}
	
	if result.Lines != 1 {
		t.Errorf("Lines = %d, want 1", result.Lines)
	}
	
	// Check that we have multiple methods
	if len(result.Methods) < 3 {
		t.Errorf("Expected at least 3 counting methods, got %d", len(result.Methods))
	}
	
	// Check approximation methods exist
	hasCharBased := false
	hasWordBased := false
	for _, method := range result.Methods {
		if strings.Contains(method.DisplayName, "Character-based") {
			hasCharBased = true
		}
		if strings.Contains(method.DisplayName, "Word-based") {
			hasWordBased = true
		}
	}

	if !hasCharBased {
		t.Error("Missing character-based approximation")
	}
	if !hasWordBased {
		t.Error("Missing word-based approximation")
	}
}

func TestGetApproximations(t *testing.T) {
	counter := NewCounter(CounterOptions{
		CharsPerToken: 4.0,
		WordsPerToken: 0.75,
	})
	
	text := "This is a test text with twenty characters and six words."
	methods := counter.getApproximations(text)
	
	if len(methods) != 3 {
		t.Errorf("Expected 3 approximation methods, got %d", len(methods))
	}
	
	// Check character-based approximation
	expectedCharTokens := len(text) / 4
	found := false
	for _, method := range methods {
		if strings.Contains(method.DisplayName, "Character-based") {
			found = true
			if method.Tokens != expectedCharTokens {
				t.Errorf("Character-based tokens = %d, want %d", method.Tokens, expectedCharTokens)
			}
		}
	}
	if !found {
		t.Error("Character-based approximation not found")
	}
}

func TestCounterOptions(t *testing.T) {
	text := "Test text"
	
	// Test with custom options
	counter := NewCounter(CounterOptions{
		CharsPerToken: 2.0,
		WordsPerToken: 0.5,
	})
	
	methods := counter.getApproximations(text)

	// Find character-based method
	for _, method := range methods {
		if strings.Contains(method.DisplayName, "÷2.0") {
			expectedTokens := len(text) / 2
			if method.Tokens != expectedTokens {
				t.Errorf("Custom char ratio: got %d tokens, want %d", method.Tokens, expectedTokens)
			}
			return
		}
	}

	t.Error("Custom character ratio not applied")
}

func BenchmarkCountWords(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = countWords(text)
	}
}

func BenchmarkCounter_Count(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)
	counter := NewCounter(CounterOptions{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = counter.Count(text, "", false)
	}
}