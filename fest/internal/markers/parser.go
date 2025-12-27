package markers

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

var markerRegex = regexp.MustCompile(`\[REPLACE:\s*([^\]]+)\]`)

// Parse extracts all REPLACE markers from content.
func Parse(content string) []Marker {
	var markers []Marker

	matches := markerRegex.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		fullMatch := content[match[0]:match[1]]
		hint := strings.TrimSpace(content[match[2]:match[3]])

		// Calculate line number (1-indexed)
		lineNumber := strings.Count(content[:match[0]], "\n") + 1

		markers = append(markers, Marker{
			FullMatch:   fullMatch,
			Hint:        hint,
			LineNumber:  lineNumber,
			StartOffset: match[0],
			EndOffset:   match[1],
		})
	}

	return markers
}

// ParseFile reads a file and extracts all REPLACE markers.
func ParseFile(ctx context.Context, filePath string) ([]Marker, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("markers.ParseFile")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading file").
			WithCode(errors.ErrCodeIO).
			WithField("path", filePath)
	}

	return Parse(string(content)), nil
}

// ExtractHints returns just the hint strings from a slice of markers.
func ExtractHints(markers []Marker) []string {
	hints := make([]string, len(markers))
	for i, m := range markers {
		hints[i] = m.Hint
	}
	return hints
}
