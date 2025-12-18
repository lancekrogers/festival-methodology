// Package understand provides embedded content for the fest understand command.
package understand

import "embed"

//go:embed overview.txt rules.txt templates.txt structure.txt methodology.txt workflow.txt tasks.txt
//go:embed scaffolds/*.txt
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

// LoadScaffold reads a scaffold file from the scaffolds directory.
func LoadScaffold(name string) string {
	return Load("scaffolds/" + name + ".txt")
}
