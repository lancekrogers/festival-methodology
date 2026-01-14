// Package intro provides embedded content for the fest intro command.
package intro

import "embed"

//go:embed intro.txt workflows.txt
var Content embed.FS

// Load reads an embedded file and returns its content as a string.
// Returns empty string if file not found.
func Load(name string) string {
	content, err := Content.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(content)
}
