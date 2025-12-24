package commands

import "github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"

func init() {
	// Register command execution hooks so tui package can call them
	// without importing commands directly (which would create a cycle).
	shared.RunInit = runInitHook
	shared.RunCreateFestival = runCreateFestivalHook
	shared.RunCreatePhase = runCreatePhaseHook
	shared.RunCreateSequence = runCreateSequenceHook
	shared.RunCreateTask = runCreateTaskHook
	shared.RunApply = runApplyHook
}

// Hook implementations that convert shared.XxxOpts to internal XxxOptions

func runInitHook(path string, opts *shared.InitOpts) error {
	return RunInit(path, &InitOptions{
		From:        opts.From,
		Minimal:     opts.Minimal,
		NoChecksums: opts.NoChecksums,
	})
}

func runCreateFestivalHook(opts *shared.CreateFestivalOpts) error {
	return RunCreateFestival(&CreateFestivalOptions{
		Name:       opts.Name,
		Goal:       opts.Goal,
		Tags:       opts.Tags,
		VarsFile:   opts.VarsFile,
		JSONOutput: opts.JSONOutput,
		Dest:       opts.Dest,
	})
}

func runCreatePhaseHook(opts *shared.CreatePhaseOpts) error {
	return RunCreatePhase(&CreatePhaseOptions{
		After:      opts.After,
		Name:       opts.Name,
		PhaseType:  opts.PhaseType,
		Path:       opts.Path,
		VarsFile:   opts.VarsFile,
		JSONOutput: opts.JSONOutput,
	})
}

func runCreateSequenceHook(opts *shared.CreateSequenceOpts) error {
	return RunCreateSequence(&CreateSequenceOptions{
		After:      opts.After,
		Name:       opts.Name,
		Path:       opts.Path,
		VarsFile:   opts.VarsFile,
		JSONOutput: opts.JSONOutput,
	})
}

func runCreateTaskHook(opts *shared.CreateTaskOpts) error {
	return RunCreateTask(&CreateTaskOptions{
		After:      opts.After,
		Names:      opts.Names,
		Path:       opts.Path,
		VarsFile:   opts.VarsFile,
		JSONOutput: opts.JSONOutput,
	})
}

func runApplyHook(opts *shared.ApplyOpts) error {
	return RunApply(&ApplyOptions{
		TemplateID:   opts.TemplateID,
		TemplatePath: opts.TemplatePath,
		DestPath:     opts.DestPath,
		VarsFile:     opts.VarsFile,
		JSONOutput:   opts.JSONOutput,
	})
}
