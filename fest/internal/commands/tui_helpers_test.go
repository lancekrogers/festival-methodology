package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIsDigits(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid single digit", "5", true},
		{"valid multiple digits", "123", true},
		{"valid with leading zero", "007", true},
		{"empty string", "", false},
		{"contains letter", "12a", false},
		{"starts with letter", "a12", false},
		{"contains space", "1 2", false},
		{"negative number", "-1", false},
		{"decimal", "1.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDigits(tt.input)
			if got != tt.want {
				t.Errorf("isDigits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple text", "My Festival", "my-festival"},
		{"uppercase", "TESTING", "testing"},
		{"special chars", "Hello! World?", "hello-world"},
		{"multiple spaces", "one   two", "one-two"},
		{"leading trailing spaces", "  spaced  ", "spaced"},
		{"numbers preserved", "Phase 01", "phase-01"},
		{"empty string", "", "festival"},
		{"only special chars", "!@#$%", "festival"},
		{"underscores to dashes", "hello_world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAtoiDefault(t *testing.T) {
	tests := []struct {
		name  string
		input string
		def   int
		want  int
	}{
		{"valid number", "42", 0, 42},
		{"valid with spaces", "  42  ", 0, 42},
		{"zero", "0", 99, 0},
		{"invalid returns default", "abc", 99, 99},
		{"empty returns default", "", 5, 5},
		{"negative number", "-10", 0, -10},
		{"partial number", "12abc", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := atoiDefault(tt.input, tt.def)
			if got != tt.want {
				t.Errorf("atoiDefault(%q, %d) = %d, want %d", tt.input, tt.def, got, tt.want)
			}
		})
	}
}

func TestIsPhaseDirPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid 3-digit", "/festival/001_PLANNING", true},
		{"valid 3-digit with longer name", "/festival/002_IMPLEMENTATION_PHASE", true},
		{"only 2 digits", "/festival/01_PLANNING", false},
		{"4 digits", "/festival/0001_PLANNING", false},
		{"no underscore", "/festival/001PLANNING", false},
		{"no name after underscore", "/festival/001_", false}, // regex requires at least one char after underscore
		{"just name no number", "/festival/PLANNING", false},
		{"nested path", "/path/to/festival/003_REVIEW", true},
		{"root path check", "001_PLANNING", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPhaseDirPath(tt.path)
			if got != tt.want {
				t.Errorf("isPhaseDirPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsSequenceDirPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid 2-digit", "/phase/01_requirements", true},
		{"valid 2-digit with longer name", "/phase/02_design_review", true},
		{"3 digits (phase pattern)", "/phase/001_requirements", false},
		{"single digit", "/phase/1_requirements", false},
		{"no underscore", "/phase/01requirements", false},
		{"nested path", "/festival/001_PLANNING/03_implementation", true},
		{"root path check", "05_testing", true},
		{"with .md extension (task)", "01_task.md", true}, // regex matches, but shouldn't be dir
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSequenceDirPath(tt.path)
			if got != tt.want {
				t.Errorf("isSequenceDirPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing file", testFile, true},
		{"existing directory", tmpDir, true},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.txt"), false},
		{"non-existent nested", filepath.Join(tmpDir, "a", "b", "c"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exists(tt.path)
			if got != tt.want {
				t.Errorf("exists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestListPhaseDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phase directories
	phases := []string{"001_PLANNING", "002_IMPLEMENT", "003_REVIEW"}
	for _, p := range phases {
		if err := os.MkdirAll(filepath.Join(tmpDir, p), 0755); err != nil {
			t.Fatalf("failed to create phase dir: %v", err)
		}
	}

	// Create non-phase items
	if err := os.MkdirAll(filepath.Join(tmpDir, "not_a_phase"), 0755); err != nil {
		t.Fatalf("failed to create non-phase dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "001_file.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	t.Run("returns only phase directories sorted", func(t *testing.T) {
		got := listPhaseDirs(tmpDir)
		if len(got) != 3 {
			t.Errorf("listPhaseDirs returned %d items, want 3", len(got))
		}
		for i, want := range phases {
			if i < len(got) && got[i] != want {
				t.Errorf("listPhaseDirs[%d] = %q, want %q", i, got[i], want)
			}
		}
	})

	t.Run("empty directory returns nil", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.MkdirAll(emptyDir, 0755)
		got := listPhaseDirs(emptyDir)
		if got != nil && len(got) != 0 {
			t.Errorf("listPhaseDirs on empty dir returned %v, want nil or empty", got)
		}
	})

	t.Run("non-existent directory returns nil", func(t *testing.T) {
		got := listPhaseDirs(filepath.Join(tmpDir, "nonexistent"))
		if got != nil {
			t.Errorf("listPhaseDirs on nonexistent dir returned %v, want nil", got)
		}
	})
}

func TestListSequenceDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sequence directories
	sequences := []string{"01_requirements", "02_design", "03_implementation"}
	for _, s := range sequences {
		if err := os.MkdirAll(filepath.Join(tmpDir, s), 0755); err != nil {
			t.Fatalf("failed to create sequence dir: %v", err)
		}
	}

	// Create non-sequence items
	if err := os.MkdirAll(filepath.Join(tmpDir, "not_a_sequence"), 0755); err != nil {
		t.Fatalf("failed to create non-sequence dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "01_task.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create task file: %v", err)
	}

	t.Run("returns only sequence directories sorted", func(t *testing.T) {
		got := listSequenceDirs(tmpDir)
		if len(got) != 3 {
			t.Errorf("listSequenceDirs returned %d items, want 3", len(got))
		}
		for i, want := range sequences {
			if i < len(got) && got[i] != want {
				t.Errorf("listSequenceDirs[%d] = %q, want %q", i, got[i], want)
			}
		}
	})

	t.Run("empty directory returns nil or empty", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.MkdirAll(emptyDir, 0755)
		got := listSequenceDirs(emptyDir)
		if got != nil && len(got) != 0 {
			t.Errorf("listSequenceDirs on empty dir returned %v, want nil or empty", got)
		}
	})
}

func TestFindFestivalDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure
	festivalDir := tmpDir
	phaseDir := filepath.Join(festivalDir, "001_PLANNING")
	sequenceDir := filepath.Join(phaseDir, "01_requirements")

	if err := os.MkdirAll(sequenceDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create FESTIVAL_OVERVIEW.md to make it look like a festival
	if err := os.WriteFile(filepath.Join(festivalDir, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create festival file: %v", err)
	}

	tests := []struct {
		name string
		cwd  string
		want string
	}{
		{"from phase dir returns parent", phaseDir, festivalDir},
		{"from festival dir returns itself", festivalDir, festivalDir},
		// Note: from sequence dir the function checks if cwd is a phase, which it's not (2-digit),
		// then checks if it looks like a festival (no phases directly), then checks parent
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findFestivalDir(tt.cwd)
			if got != tt.want {
				t.Errorf("findFestivalDir(%q) = %q, want %q", tt.cwd, got, tt.want)
			}
		})
	}
}

func TestLooksLikeFestivalDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a festival-like directory with FESTIVAL_OVERVIEW.md
	festivalWithOverview := filepath.Join(tmpDir, "with_overview")
	os.MkdirAll(festivalWithOverview, 0755)
	os.WriteFile(filepath.Join(festivalWithOverview, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)

	// Create a festival-like directory with FESTIVAL_GOAL.md
	festivalWithGoal := filepath.Join(tmpDir, "with_goal")
	os.MkdirAll(festivalWithGoal, 0755)
	os.WriteFile(filepath.Join(festivalWithGoal, "FESTIVAL_GOAL.md"), []byte("test"), 0644)

	// Create a directory with phase directories
	festivalWithPhases := filepath.Join(tmpDir, "with_phases")
	os.MkdirAll(filepath.Join(festivalWithPhases, "001_PLANNING"), 0755)

	// Create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(emptyDir, 0755)

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"has FESTIVAL_OVERVIEW.md", festivalWithOverview, true},
		{"has FESTIVAL_GOAL.md", festivalWithGoal, true},
		{"has phase directories", festivalWithPhases, true},
		{"empty directory", emptyDir, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeFestivalDir(tt.dir)
			if got != tt.want {
				t.Errorf("looksLikeFestivalDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestResolvePhaseDirInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure
	festivalDir := tmpDir
	os.MkdirAll(filepath.Join(festivalDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(festivalDir, "002_IMPLEMENT"), 0755)
	os.MkdirAll(filepath.Join(festivalDir, "003_REVIEW"), 0755)
	os.WriteFile(filepath.Join(festivalDir, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)

	phaseDir := filepath.Join(festivalDir, "002_IMPLEMENT")

	tests := []struct {
		name      string
		input     string
		cwd       string
		want      string
		wantError bool
	}{
		{
			name:      "numeric shortcut 02",
			input:     "02",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "numeric shortcut 2",
			input:     "2",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "3-digit numeric",
			input:     "002",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "name match IMPLEMENT",
			input:     "IMPLEMENT",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "case insensitive name",
			input:     "implement",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "exact full name",
			input:     "002_IMPLEMENT",
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "002_IMPLEMENT"),
			wantError: false,
		},
		{
			name:      "empty input from phase dir",
			input:     "",
			cwd:       phaseDir,
			want:      phaseDir,
			wantError: false,
		},
		{
			name:      "dot from phase dir",
			input:     ".",
			cwd:       phaseDir,
			want:      phaseDir,
			wantError: false,
		},
		{
			name:      "empty input from festival dir errors",
			input:     "",
			cwd:       festivalDir,
			want:      "",
			wantError: true,
		},
		{
			name:      "non-existent phase",
			input:     "99",
			cwd:       festivalDir,
			want:      "",
			wantError: true,
		},
		{
			name:      "non-existent name",
			input:     "NONEXISTENT",
			cwd:       festivalDir,
			want:      "",
			wantError: true,
		},
		{
			name:      "absolute path",
			input:     filepath.Join(festivalDir, "001_PLANNING"),
			cwd:       festivalDir,
			want:      filepath.Join(festivalDir, "001_PLANNING"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePhaseDirInput(tt.input, tt.cwd)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("resolvePhaseDirInput(%q, %q) = %q, want %q", tt.input, tt.cwd, got, tt.want)
			}
		})
	}
}

func TestResolveSequenceDirInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure
	festivalDir := tmpDir
	phaseDir := filepath.Join(festivalDir, "001_PLANNING")
	os.MkdirAll(filepath.Join(phaseDir, "01_requirements"), 0755)
	os.MkdirAll(filepath.Join(phaseDir, "02_design"), 0755)
	os.MkdirAll(filepath.Join(phaseDir, "03_implementation"), 0755)
	os.WriteFile(filepath.Join(festivalDir, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)

	sequenceDir := filepath.Join(phaseDir, "02_design")

	tests := []struct {
		name      string
		input     string
		cwd       string
		want      string
		wantError bool
	}{
		{
			name:      "numeric shortcut 02",
			input:     "02",
			cwd:       phaseDir,
			want:      filepath.Join(phaseDir, "02_design"),
			wantError: false,
		},
		{
			name:      "numeric shortcut 2",
			input:     "2",
			cwd:       phaseDir,
			want:      filepath.Join(phaseDir, "02_design"),
			wantError: false,
		},
		{
			name:      "name match design",
			input:     "design",
			cwd:       phaseDir,
			want:      filepath.Join(phaseDir, "02_design"),
			wantError: false,
		},
		{
			name:      "exact full name",
			input:     "02_design",
			cwd:       phaseDir,
			want:      filepath.Join(phaseDir, "02_design"),
			wantError: false,
		},
		{
			name:      "empty input from sequence dir",
			input:     "",
			cwd:       sequenceDir,
			want:      sequenceDir,
			wantError: false,
		},
		{
			name:      "dot from sequence dir",
			input:     ".",
			cwd:       sequenceDir,
			want:      sequenceDir,
			wantError: false,
		},
		{
			name:      "empty input from phase dir errors",
			input:     "",
			cwd:       phaseDir,
			want:      "",
			wantError: true,
		},
		{
			name:      "non-existent sequence",
			input:     "99",
			cwd:       phaseDir,
			want:      "",
			wantError: true,
		},
		{
			name:      "from sequence dir resolves sibling",
			input:     "01",
			cwd:       sequenceDir,
			want:      filepath.Join(phaseDir, "01_requirements"),
			wantError: false,
		},
		{
			name:      "absolute path",
			input:     filepath.Join(phaseDir, "01_requirements"),
			cwd:       phaseDir,
			want:      filepath.Join(phaseDir, "01_requirements"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSequenceDirInput(tt.input, tt.cwd)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("resolveSequenceDirInput(%q, %q) = %q, want %q", tt.input, tt.cwd, got, tt.want)
			}
		})
	}
}

func TestNextPhaseAfter(t *testing.T) {
	t.Run("empty festival returns 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		got := nextPhaseAfter(tmpDir)
		if got != 0 {
			t.Errorf("nextPhaseAfter on empty dir = %d, want 0", got)
		}
	})

	t.Run("single phase returns its number", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
		got := nextPhaseAfter(tmpDir)
		if got != 1 {
			t.Errorf("nextPhaseAfter with one phase = %d, want 1", got)
		}
	})

	t.Run("multiple phases returns max", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "003_REVIEW"), 0755)
		got := nextPhaseAfter(tmpDir)
		if got != 3 {
			t.Errorf("nextPhaseAfter with three phases = %d, want 3", got)
		}
	})

	t.Run("gaps in numbering returns max", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "005_DEPLOY"), 0755)
		got := nextPhaseAfter(tmpDir)
		if got != 5 {
			t.Errorf("nextPhaseAfter with gap = %d, want 5", got)
		}
	})
}

func TestNextSequenceAfter(t *testing.T) {
	t.Run("empty phase returns 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		got := nextSequenceAfter(tmpDir)
		if got != 0 {
			t.Errorf("nextSequenceAfter on empty dir = %d, want 0", got)
		}
	})

	t.Run("multiple sequences returns max", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "01_req"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "02_design"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "03_impl"), 0755)
		got := nextSequenceAfter(tmpDir)
		if got != 3 {
			t.Errorf("nextSequenceAfter = %d, want 3", got)
		}
	})
}

func TestNextTaskAfter(t *testing.T) {
	t.Run("empty sequence returns 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		got := nextTaskAfter(tmpDir)
		if got != 0 {
			t.Errorf("nextTaskAfter on empty dir = %d, want 0", got)
		}
	})

	t.Run("multiple tasks returns max", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "01_task_a.md"), []byte("task"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "02_task_b.md"), []byte("task"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "03_task_c.md"), []byte("task"), 0644)
		got := nextTaskAfter(tmpDir)
		if got != 3 {
			t.Errorf("nextTaskAfter = %d, want 3", got)
		}
	})
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "removes duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "filters empty strings",
			input: []string{"a", "", "b", ""},
			want:  []string{"a", "b"},
		},
		{
			name:  "preserves order",
			input: []string{"z", "a", "m"},
			want:  []string{"z", "a", "m"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "all empty strings",
			input: []string{"", "", ""},
			want:  []string{},
		},
		{
			name:  "nil input",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueStrings(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("uniqueStrings(%v) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("uniqueStrings(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWriteTempVarsFile(t *testing.T) {
	t.Run("empty map returns empty string", func(t *testing.T) {
		path, err := writeTempVarsFile(map[string]interface{}{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path != "" {
			t.Errorf("writeTempVarsFile({}) = %q, want empty string", path)
		}
	})

	t.Run("nil map returns empty string", func(t *testing.T) {
		path, err := writeTempVarsFile(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path != "" {
			t.Errorf("writeTempVarsFile(nil) = %q, want empty string", path)
		}
	})

	t.Run("writes valid JSON", func(t *testing.T) {
		vars := map[string]interface{}{
			"name":   "test-festival",
			"count":  42,
			"active": true,
		}

		path, err := writeTempVarsFile(vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer os.Remove(path)

		if path == "" {
			t.Fatal("expected non-empty path")
		}

		// Read and verify JSON
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		var got map[string]interface{}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		if got["name"] != "test-festival" {
			t.Errorf("name = %v, want test-festival", got["name"])
		}
		if got["count"] != float64(42) { // JSON numbers are float64
			t.Errorf("count = %v, want 42", got["count"])
		}
		if got["active"] != true {
			t.Errorf("active = %v, want true", got["active"])
		}
	})
}

func TestDefaultFestivalTemplatePaths(t *testing.T) {
	tmplRoot := "/templates"
	got := defaultFestivalTemplatePaths(tmplRoot)

	expected := []string{
		"/templates/FESTIVAL_OVERVIEW_TEMPLATE.md",
		"/templates/FESTIVAL_GOAL_TEMPLATE.md",
		"/templates/FESTIVAL_RULES_TEMPLATE.md",
		"/templates/FESTIVAL_TODO_TEMPLATE.md",
	}

	if len(got) != len(expected) {
		t.Errorf("defaultFestivalTemplatePaths returned %d paths, want %d", len(got), len(expected))
		return
	}

	for i, want := range expected {
		if got[i] != want {
			t.Errorf("defaultFestivalTemplatePaths[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestCollectRequiredVars(t *testing.T) {
	// Note: This test is limited because it depends on the template loader
	// and actual template files. We test the edge cases we can.

	t.Run("non-existent paths returns empty", func(t *testing.T) {
		got := collectRequiredVars("/nonexistent", []string{
			"/nonexistent/a.md",
			"/nonexistent/b.md",
		})
		if len(got) != 0 {
			t.Errorf("collectRequiredVars with missing files returned %v, want empty", got)
		}
	})

	t.Run("empty paths returns empty", func(t *testing.T) {
		got := collectRequiredVars("/any", []string{})
		if len(got) != 0 {
			t.Errorf("collectRequiredVars with empty paths returned %v, want empty", got)
		}
	})
}
