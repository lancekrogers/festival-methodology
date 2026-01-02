package festival

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// TestCreateSequence_DirectoryCreation tests that creating a sequence
// results in a directory being created (not a file)
func TestCreateSequence_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up phase directory structure
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatalf("failed to create phase dir: %v", err)
	}

	// Use renumberer directly (what create sequence uses internally)
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertSequence(context.Background(), phaseDir, 0, "requirements")
	if err != nil {
		t.Fatalf("InsertSequence failed: %v", err)
	}

	// Verify directory was created (not a file)
	seqPath := filepath.Join(phaseDir, "01_requirements")
	info, err := os.Stat(seqPath)
	if err != nil {
		t.Fatalf("expected sequence directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected 01_requirements to be a directory, got file")
	}
}

// TestCreateTask_FileCreation tests that creating a task
// creates the parent directory but doesn't write file content
func TestCreateTask_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up sequence directory structure
	seqDir := filepath.Join(tmpDir, "01_requirements")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("failed to create sequence dir: %v", err)
	}

	// Use renumberer directly (what create task uses internally)
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertTask(context.Background(), seqDir, 0, "define_requirements")
	if err != nil {
		t.Fatalf("InsertTask failed: %v", err)
	}

	// The renumberer ensures parent exists but doesn't create .md file content
	// (the create command is responsible for writing template content)
	taskPath := filepath.Join(seqDir, "01_define_requirements.md")

	// Check parent directory exists
	if _, err := os.Stat(seqDir); err != nil {
		t.Fatalf("expected parent directory to exist: %v", err)
	}

	// Task file may or may not exist (renumberer behavior)
	// but if it exists, it should be empty (no template content)
	if info, err := os.Stat(taskPath); err == nil {
		if info.IsDir() {
			t.Error("expected task to be a file, not a directory")
		}
		// If file exists, verify it's empty (no content written by renumberer)
		content, _ := os.ReadFile(taskPath)
		if len(content) > 0 {
			t.Errorf("renumberer should not write content to task file, got: %q", content)
		}
	}
}

// TestCreateSequence_NoIsDirectoryError tests that creating a sequence
// doesn't cause "is a directory" errors when the directory already exists
func TestCreateSequence_NoIsDirectoryError(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up phase directory with existing sequence
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	existingSeq := filepath.Join(phaseDir, "01_existing")
	if err := os.MkdirAll(existingSeq, 0755); err != nil {
		t.Fatalf("failed to create existing sequence: %v", err)
	}

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert new sequence at beginning (should shift existing)
	err := ren.InsertSequence(context.Background(), phaseDir, 0, "new_sequence")
	if err != nil {
		t.Fatalf("InsertSequence failed with error: %v", err)
	}

	// Verify both sequences exist as directories
	newSeq := filepath.Join(phaseDir, "01_new_sequence")
	shiftedSeq := filepath.Join(phaseDir, "02_existing")

	if info, err := os.Stat(newSeq); err != nil || !info.IsDir() {
		t.Error("expected 01_new_sequence directory")
	}
	if info, err := os.Stat(shiftedSeq); err != nil || !info.IsDir() {
		t.Error("expected 02_existing directory (renumbered)")
	}
}

// TestCreateTask_NoIsDirectoryError tests that creating a task
// with existing tasks doesn't cause errors
func TestCreateTask_NoIsDirectoryError(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up sequence directory with existing task
	seqDir := tmpDir
	if err := os.WriteFile(filepath.Join(seqDir, "01_existing.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create existing task: %v", err)
	}

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert new task at beginning (should shift existing)
	err := ren.InsertTask(context.Background(), seqDir, 0, "new_task")
	if err != nil {
		t.Fatalf("InsertTask failed with error: %v", err)
	}

	// Verify renumbering happened correctly
	shiftedTask := filepath.Join(seqDir, "02_existing.md")
	if _, err := os.Stat(shiftedTask); err != nil {
		t.Error("expected 02_existing.md (renumbered from 01)")
	}

	// Verify content was preserved in renamed file
	content, err := os.ReadFile(shiftedTask)
	if err != nil {
		t.Fatalf("failed to read shifted task: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("shifted task content = %q, want %q", content, "content")
	}
}

// TestCreatePhase_DirectoryCreation tests that creating a phase
// results in a directory being created
func TestCreatePhase_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertPhase(context.Background(), tmpDir, 0, "PLANNING")
	if err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Verify directory was created
	phasePath := filepath.Join(tmpDir, "001_PLANNING")
	info, err := os.Stat(phasePath)
	if err != nil {
		t.Fatalf("expected phase directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected 001_PLANNING to be a directory, got file")
	}
}

// TestCreate_ChainedOperations tests multiple create operations in sequence
func TestCreate_ChainedOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure progressively
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Create phases
	ctx := context.Background()
	if err := ren.InsertPhase(ctx, tmpDir, 0, "PLANNING"); err != nil {
		t.Fatalf("InsertPhase PLANNING failed: %v", err)
	}
	if err := ren.InsertPhase(ctx, tmpDir, 1, "IMPLEMENT"); err != nil {
		t.Fatalf("InsertPhase IMPLEMENT failed: %v", err)
	}

	// Create sequences in first phase
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	if err := ren.InsertSequence(ctx, phaseDir, 0, "requirements"); err != nil {
		t.Fatalf("InsertSequence requirements failed: %v", err)
	}
	if err := ren.InsertSequence(ctx, phaseDir, 1, "design"); err != nil {
		t.Fatalf("InsertSequence design failed: %v", err)
	}

	// Create tasks in first sequence
	seqDir := filepath.Join(phaseDir, "01_requirements")
	if err := ren.InsertTask(ctx, seqDir, 0, "gather_info"); err != nil {
		t.Fatalf("InsertTask gather_info failed: %v", err)
	}
	if err := ren.InsertTask(ctx, seqDir, 1, "analyze"); err != nil {
		t.Fatalf("InsertTask analyze failed: %v", err)
	}

	// Verify complete structure
	expectedDirs := []string{
		"001_PLANNING",
		"002_IMPLEMENT",
		"001_PLANNING/01_requirements",
		"001_PLANNING/02_design",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}
}

// TestCreateOptions_DefaultAfterZero verifies that after=0 creates at position 1
func TestCreateOptions_DefaultAfterZero(t *testing.T) {
	tmpDir := t.TempDir()

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Create with after=0 (default)
	if err := ren.InsertPhase(context.Background(), tmpDir, 0, "FIRST"); err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Should create at position 1 (001_)
	if _, err := os.Stat(filepath.Join(tmpDir, "001_FIRST")); err != nil {
		t.Error("expected 001_FIRST to exist")
	}
}

// TestCreateOptions_InsertInMiddle verifies inserting in the middle of existing elements
func TestCreateOptions_InsertInMiddle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial phases
	os.MkdirAll(filepath.Join(tmpDir, "001_FIRST"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_THIRD"), 0755)

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert after position 1
	if err := ren.InsertPhase(context.Background(), tmpDir, 1, "SECOND"); err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Verify structure
	expected := []string{"001_FIRST", "002_SECOND", "003_THIRD"}
	for _, name := range expected {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); err != nil {
			t.Errorf("expected %s to exist", name)
		}
	}
}

// TestCreateFestival_GatesDirectory tests that festival creation creates gates directory
func TestCreateFestival_GatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the festivals directory structure with templates
	festivalsDir := filepath.Join(tmpDir, "festivals")
	festivalMetaDir := filepath.Join(festivalsDir, ".festival")
	templatesDir := filepath.Join(festivalMetaDir, "templates")
	gatesTemplatesDir := filepath.Join(templatesDir, "gates")
	if err := os.MkdirAll(gatesTemplatesDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create minimal gate templates
	gateTemplates := []string{
		"QUALITY_GATE_TESTING.md",
		"QUALITY_GATE_REVIEW.md",
		"QUALITY_GATE_ITERATE.md",
		"QUALITY_GATE_COMMIT.md",
	}
	for _, tmpl := range gateTemplates {
		content := "# " + tmpl + "\n\nGate template content."
		if err := os.WriteFile(filepath.Join(gatesTemplatesDir, tmpl), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", tmpl, err)
		}
	}

	// Also create core templates to satisfy festival creation
	coreTemplates := []string{
		"FESTIVAL_OVERVIEW_TEMPLATE.md",
		"FESTIVAL_GOAL_TEMPLATE.md",
	}
	for _, tmpl := range coreTemplates {
		content := "# {{.festival_name}}\n"
		if err := os.WriteFile(filepath.Join(templatesDir, tmpl), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", tmpl, err)
		}
	}

	// Change working directory temporarily
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(festivalsDir)

	// Run create festival
	opts := &CreateFestivalOptions{
		Name:        "test-festival",
		Goal:        "Test goal",
		SkipMarkers: true,
		Dest:        "active",
	}

	err := RunCreateFestival(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunCreateFestival failed: %v", err)
	}

	// Find the created festival directory (now includes ID suffix)
	activeDir := filepath.Join(festivalsDir, "active")
	entries, err := os.ReadDir(activeDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 entry in active/: %v", err)
	}
	festivalDir := filepath.Join(activeDir, entries[0].Name())
	gatesDir := filepath.Join(festivalDir, "gates")

	info, err := os.Stat(gatesDir)
	if err != nil {
		t.Fatalf("expected gates directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected gates to be a directory")
	}

	// Verify gate templates were copied
	for _, tmpl := range gateTemplates {
		gatePath := filepath.Join(gatesDir, tmpl)
		if _, err := os.Stat(gatePath); err != nil {
			t.Errorf("expected gate template %s to exist: %v", tmpl, err)
		}
	}
}

// TestCreateFestival_FestYAMLGenerated tests that fest.yaml is generated with gates config
func TestCreateFestival_FestYAMLGenerated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the festivals directory structure with templates
	festivalsDir := filepath.Join(tmpDir, "festivals")
	festivalMetaDir := filepath.Join(festivalsDir, ".festival")
	templatesDir := filepath.Join(festivalMetaDir, "templates")
	gatesTemplatesDir := filepath.Join(templatesDir, "gates")
	if err := os.MkdirAll(gatesTemplatesDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}

	// Create minimal gate templates
	for _, tmpl := range []string{"QUALITY_GATE_TESTING.md", "QUALITY_GATE_REVIEW.md", "QUALITY_GATE_ITERATE.md", "QUALITY_GATE_COMMIT.md"} {
		if err := os.WriteFile(filepath.Join(gatesTemplatesDir, tmpl), []byte("# Gate"), 0644); err != nil {
			t.Fatalf("failed to create template: %v", err)
		}
	}

	// Also create core templates
	for _, tmpl := range []string{"FESTIVAL_OVERVIEW_TEMPLATE.md", "FESTIVAL_GOAL_TEMPLATE.md"} {
		if err := os.WriteFile(filepath.Join(templatesDir, tmpl), []byte("# Template"), 0644); err != nil {
			t.Fatalf("failed to create template: %v", err)
		}
	}

	// Change working directory temporarily
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(festivalsDir)

	// Run create festival
	opts := &CreateFestivalOptions{
		Name:        "gates-test",
		Goal:        "Test gates configuration",
		SkipMarkers: true,
		Dest:        "active",
	}

	err := RunCreateFestival(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunCreateFestival failed: %v", err)
	}

	// Find the created festival directory (now includes ID suffix)
	activeDir := filepath.Join(festivalsDir, "active")
	entries, err := os.ReadDir(activeDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 entry in active/: %v", err)
	}
	festivalDir := filepath.Join(activeDir, entries[0].Name())
	festYAMLPath := filepath.Join(festivalDir, "fest.yaml")

	if _, err := os.Stat(festYAMLPath); err != nil {
		t.Fatalf("expected fest.yaml to exist: %v", err)
	}

	// Read and verify content has gates/ prefix
	content, err := os.ReadFile(festYAMLPath)
	if err != nil {
		t.Fatalf("failed to read fest.yaml: %v", err)
	}
	contentStr := string(content)

	// Check that gates/ prefix is used in template paths
	if !contains(contentStr, "gates/QUALITY_GATE_TESTING") {
		t.Error("fest.yaml should contain gates/QUALITY_GATE_TESTING")
	}
	if !contains(contentStr, "gates/QUALITY_GATE_REVIEW") {
		t.Error("fest.yaml should contain gates/QUALITY_GATE_REVIEW")
	}
	if !contains(contentStr, "gates/QUALITY_GATE_ITERATE") {
		t.Error("fest.yaml should contain gates/QUALITY_GATE_ITERATE")
	}
	if !contains(contentStr, "gates/QUALITY_GATE_COMMIT") {
		t.Error("fest.yaml should contain gates/QUALITY_GATE_COMMIT")
	}

	// Verify quality_gates.enabled is true
	if !contains(contentStr, "enabled: true") {
		t.Error("fest.yaml should have quality_gates.enabled: true")
	}
}

// TestCreateFestival_GatesConfigHasCorrectStructure verifies the generated config
func TestCreateFestival_GatesConfigHasCorrectStructure(t *testing.T) {
	cfg := DefaultFestivalGatesConfig()

	// Verify quality gates are enabled
	if !cfg.QualityGates.Enabled {
		t.Error("quality gates should be enabled by default")
	}

	// Verify we have 4 default gates
	if len(cfg.QualityGates.Tasks) != 4 {
		t.Errorf("expected 4 quality gate tasks, got %d", len(cfg.QualityGates.Tasks))
	}

	// Verify all gates use gates/ prefix
	expectedTemplates := map[string]bool{
		"gates/QUALITY_GATE_TESTING": false,
		"gates/QUALITY_GATE_REVIEW":  false,
		"gates/QUALITY_GATE_ITERATE": false,
		"gates/QUALITY_GATE_COMMIT":  false,
	}

	for _, task := range cfg.QualityGates.Tasks {
		if _, ok := expectedTemplates[task.Template]; !ok {
			t.Errorf("unexpected template path: %s", task.Template)
		} else {
			expectedTemplates[task.Template] = true
		}
		if !task.Enabled {
			t.Errorf("expected task %s to be enabled", task.ID)
		}
	}

	for tmpl, found := range expectedTemplates {
		if !found {
			t.Errorf("expected template %s not found in config", tmpl)
		}
	}
}

// contains checks if substr is in s (simple substring check)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
