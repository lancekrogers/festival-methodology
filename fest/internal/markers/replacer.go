package markers

import (
	"context"
	"os"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Replace applies marker values to content, returning the modified content.
// Replaces in reverse order to preserve character offsets.
func Replace(content string, values []MarkerValue) string {
	result := content

	// Replace in reverse order to preserve offsets
	for i := len(values) - 1; i >= 0; i-- {
		mv := values[i]
		result = strings.Replace(result, mv.Marker.FullMatch, mv.Value, 1)
	}

	return result
}

// ReplaceInFile reads a file, applies marker values, and writes it back.
func ReplaceInFile(ctx context.Context, filePath string, values []MarkerValue) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("markers.ReplaceInFile")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "reading file").
			WithCode(errors.ErrCodeIO).
			WithField("path", filePath)
	}

	newContent := Replace(string(content), values)

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return errors.Wrap(err, "writing file").
			WithCode(errors.ErrCodeIO).
			WithField("path", filePath)
	}

	return nil
}

// ApplyInput matches input values to markers and creates MarkerValue slice.
// Values not found in input keep the original marker text (unfilled).
func ApplyInput(markers []Marker, input map[string]string) []MarkerValue {
	values := make([]MarkerValue, len(markers))

	for i, marker := range markers {
		value := marker.FullMatch // Default: keep original (unfilled)

		if v, ok := input[marker.Hint]; ok && v != "" {
			value = v
		}

		values[i] = MarkerValue{
			Marker: marker,
			Value:  value,
		}
	}

	return values
}

// ComputeResult analyzes marker values and returns processing statistics.
func ComputeResult(filePath string, values []MarkerValue) Result {
	result := Result{
		FilePath:     filePath,
		TotalMarkers: len(values),
	}

	for _, v := range values {
		if v.Value != v.Marker.FullMatch {
			result.FilledMarkers++
		} else {
			result.UnfilledHints = append(result.UnfilledHints, v.Marker.Hint)
		}
	}

	return result
}
