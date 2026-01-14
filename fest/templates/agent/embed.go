// Package agent provides embedded templates for agent-facing CLI output.
// All agent instruction templates live here for centralized maintainability.
package agent

import "embed"

//go:embed next/*.tmpl
//go:embed execute/*.tmpl
//go:embed validate/*.tmpl
//go:embed gates/implementation/*.tmpl
//go:embed gates/research/*.tmpl
//go:embed gates/planning/*.tmpl
//go:embed gates/review/*.tmpl
//go:embed gates/non_coding_action/*.tmpl
var Templates embed.FS
