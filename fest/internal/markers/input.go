package markers

import (
	"context"
	"encoding/json"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ParseJSON parses a JSON string into a hint→value map.
func ParseJSON(jsonStr string) (map[string]string, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var input map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		return nil, errors.Wrap(err, "parsing markers JSON").
			WithCode(errors.ErrCodeParse)
	}

	return input, nil
}

// ReadJSONFile reads a JSON file into a hint→value map.
func ReadJSONFile(ctx context.Context, filePath string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("markers.ReadJSONFile")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading markers file").
			WithCode(errors.ErrCodeIO).
			WithField("path", filePath)
	}

	var input map[string]string
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, errors.Wrap(err, "parsing markers file").
			WithCode(errors.ErrCodeParse).
			WithField("path", filePath)
	}

	return input, nil
}

// MarkersToJSON converts markers to a JSON-serializable structure for --dry-run output.
func MarkersToJSON(markers []Marker) []map[string]interface{} {
	result := make([]map[string]interface{}, len(markers))
	for i, m := range markers {
		result[i] = map[string]interface{}{
			"hint": m.Hint,
			"line": m.LineNumber,
		}
	}
	return result
}
