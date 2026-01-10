package shared

import (
	"encoding/json"
	"io"
)

// EncodeJSON writes a JSON representation of value to the writer.
func EncodeJSON(out io.Writer, value interface{}) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
