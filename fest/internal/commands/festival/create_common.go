package festival

import (
	"context"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/validator"
)

// ValidationSummary is a simplified validation result for JSON output
type ValidationSummary struct {
	OK       bool              `json:"ok"`
	Score    int               `json:"score"`
	Errors   int               `json:"errors"`
	Warnings int               `json:"warnings"`
	Issues   []validator.Issue `json:"issues,omitempty"`
}

// LoadEffectiveAgentConfig loads and merges workspace + festival agent configs
// Returns nil if no config is found (all defaults apply)
func LoadEffectiveAgentConfig(festivalsRoot, festivalPath string) *config.AgentConfig {
	var workspaceAgent *config.AgentConfig
	var festivalAgent *config.AgentConfig

	// Load workspace config
	if festivalsRoot != "" {
		workspaceCfg, err := config.LoadWorkspaceConfig(festivalsRoot)
		if err == nil && workspaceCfg != nil {
			workspaceAgent = &workspaceCfg.Agent
		}
	}

	// Load festival config
	if festivalPath != "" {
		festivalCfg, err := config.LoadFestivalConfig(festivalPath)
		if err == nil && festivalCfg != nil {
			festivalAgent = &festivalCfg.Agent
		}
	}

	return config.MergeAgentConfig(workspaceAgent, festivalAgent)
}

// RunPostCreateValidation runs validation after creation and returns a summary
func RunPostCreateValidation(ctx context.Context, festivalPath string) (*ValidationSummary, error) {
	if festivalPath == "" {
		return &ValidationSummary{OK: true, Score: 100}, nil
	}

	result, err := validator.QuickValidate(ctx, festivalPath)
	if err != nil {
		return nil, err
	}

	// Count errors and warnings
	errorCount := 0
	warningCount := 0
	for _, issue := range result.Issues {
		switch issue.Level {
		case validator.LevelError:
			errorCount++
		case validator.LevelWarning:
			warningCount++
		}
	}

	return &ValidationSummary{
		OK:       result.OK,
		Score:    result.Score,
		Errors:   errorCount,
		Warnings: warningCount,
		Issues:   result.Issues,
	}, nil
}

// ResolveFestivalPath finds the festival root from the current working directory
// Returns empty string if not inside a festival
func ResolveFestivalPath(cwd string) string {
	festivalPath, err := tpl.FindFestivalRoot(cwd)
	if err != nil {
		return ""
	}
	return festivalPath
}

// ResolveFestivalsRoot finds the festivals root from the current working directory
// Returns empty string if not inside a festivals structure
func ResolveFestivalsRoot(cwd string) string {
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return ""
	}
	return festivalsRoot
}
