package gates

import (
	"encoding/json"
	"os"
)

func emitApplyError(opts *applyOptions, err error) error {
	if opts.jsonOutput {
		return emitApplyResult(opts, applyResult{
			OK:     false,
			Action: "gates_apply",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
	}
	return err
}

func emitApplyResult(opts *applyOptions, result applyResult) error {
	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	return nil
}
