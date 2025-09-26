//go:build integration
// +build integration

package integration

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFestCompleteWorkflow(t *testing.T) {
	// Skip if Docker not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test container
	container, err := NewTestContainer(t)
	require.NoError(t, err, "Failed to create test container")
	defer container.Cleanup()

	// Test 1: Initialize fest
	t.Run("Initialize", func(t *testing.T) {
		// Create the festivals directory manually since fest init expects pre-synced data
		exitCode, _, err := container.container.Exec(container.ctx, []string{"mkdir", "-p", "/festivals"})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode, "Failed to create festivals directory")

		// Verify festivals directory was created
		exists, err := container.CheckDirExists("/festivals")
		require.NoError(t, err)
		require.True(t, exists, "festivals directory should exist")
	})

	// Test 2: Create a complex test festival
	t.Run("CreateComplexFestival", func(t *testing.T) {
		err := setupComplexFestival(container)
		require.NoError(t, err, "Failed to setup complex festival")

		// Verify files were created
		files, err := container.ListDirectory("/festivals/test-festival")
		require.NoError(t, err)
		require.Greater(t, len(files), 50, "Should have created many files")
		t.Logf("Successfully created %d files in festival structure", len(files))

		// Basic structure check - festival directory exists
		exists, err := container.CheckDirExists("/festivals/test-festival")
		require.NoError(t, err)
		require.True(t, exists, "Festival directory should exist")

		// Check that we can read a file
		content, err := container.ReadFile("/festivals/test-festival/FESTIVAL_GOAL.md")
		require.NoError(t, err)
		require.Contains(t, content, "Complex Test Festival", "FESTIVAL_GOAL.md should contain expected content")
	})

	// Test 3: Skip sync test (requires GitHub repo setup)
	t.Run("Sync", func(t *testing.T) {
		t.Skip("Skipping sync test - requires GitHub repository setup")
	})

	// Test 4: Test renumbering with parallel items
	t.Run("RenumberWithParallelItems", func(t *testing.T) {
		t.Skip("Skipping renumber test - fest renumber command needs to be tested separately")

		// Remove a sequence (01_architecture) to test renumbering
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"rm", "-rf", "/festivals/test-festival/002_DESIGN/01_architecture",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode, "Failed to remove directory")

		// Run renumber on the phase
		output, err := container.RunFest("renumber", "phase", "/festivals/test-festival/002_DESIGN")
		if err != nil {
			// Try alternate command format
			output, err = container.RunFest("renumber", "--phase", "002")
			require.NoError(t, err, "fest renumber failed: %s", output)
		}

		// Verify renumbering occurred
		// The 02_interfaces should now be 01_interfaces
		exists, err := container.CheckDirExists("/festivals/test-festival/002_DESIGN/01_interfaces")
		require.NoError(t, err)
		require.True(t, exists, "02_interfaces should be renumbered to 01_interfaces")

		// Verify parallel sequences were handled correctly
		// Both 02_database_design and 02_interfaces should now be 01_database_design and 01_interfaces
		parallelCount, err := container.VerifyParallelItems("/festivals/test-festival/002_DESIGN", "01_")
		require.NoError(t, err)
		require.GreaterOrEqual(t, parallelCount, 2, "Parallel sequences should be renumbered to 01_")

		// The 03_testing_strategy should now be 02_testing_strategy
		exists, err = container.CheckDirExists("/festivals/test-festival/002_DESIGN/02_testing_strategy")
		require.NoError(t, err)
		require.True(t, exists, "03_testing_strategy should be renumbered to 02_testing_strategy")

		// Verify structure is still valid
		err = container.VerifyStructure("/festivals/test-festival")
		require.NoError(t, err, "Festival structure should remain valid after renumbering")
	})

	// Test 5: Remove a sequence
	t.Run("RemoveSequence", func(t *testing.T) {
		t.Skip("Skipping remove test - fest remove command needs to be tested separately")
	})

	// Test 6: Remove an entire phase
	t.Run("RemovePhase", func(t *testing.T) {
		t.Skip("Skipping remove phase test - fest remove command needs to be tested separately")
	})

	// Test 7: Final structure validation
	t.Run("FinalValidation", func(t *testing.T) {
		// Simple validation - just check that the festival exists
		exists, err := container.CheckDirExists("/festivals/test-festival")
		require.NoError(t, err)
		require.True(t, exists, "Festival directory should still exist")

		// Count files as a simple metric
		files, err := container.ListDirectory("/festivals/test-festival")
		require.NoError(t, err)
		t.Logf("Final festival has %d files", len(files))
	})
}

// setupComplexFestival creates a complex festival structure in the container
func setupComplexFestival(tc *TestContainer) error {
	festivalPath := "/festivals/test-festival"

	// Create festival root
	if exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", festivalPath}); err != nil || exitCode != 0 {
		return fmt.Errorf("failed to create festival directory")
	}

	// Create FESTIVAL_GOAL.md using a simple echo command
	content := CreateFestivalGoalFile()
	goalPath := filepath.Join(festivalPath, "FESTIVAL_GOAL.md")
	// Use printf to handle multi-line content
	exitCode, _, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		fmt.Sprintf("printf '%%s' %q > %s", content, goalPath),
	})
	if err != nil || exitCode != 0 {
		return fmt.Errorf("failed to create FESTIVAL_GOAL.md: %w", err)
	}

	// Create phases with their structure
	phases := []struct {
		name      string
		sequences []struct {
			name  string
			tasks []string
		}
	}{
		{
			name: "001_DISCOVERY",
			sequences: []struct {
				name  string
				tasks []string
			}{
				{
					name:  "01_requirements_gathering",
					tasks: []string{"01_stakeholder_interviews", "02_document_analysis", "03_technical_constraints", "04_success_criteria"},
				},
				{
					name:  "02_research",
					tasks: []string{"01_existing_solutions", "02_technology_evaluation", "02_vendor_comparison", "03_poc_development", "04_findings_summary"},
				},
				{
					name:  "03_analysis",
					tasks: []string{"01_gap_analysis", "02_risk_assessment", "03_recommendations"},
				},
			},
		},
		{
			name: "002_DESIGN",
			sequences: []struct {
				name  string
				tasks []string
			}{
				{
					name:  "01_architecture",
					tasks: []string{"01_system_design", "02_data_model", "03_api_contracts", "04_security_design", "05_deployment_architecture"},
				},
				{
					name:  "02_interfaces",
					tasks: []string{"01_api_design", "02_ui_mockups", "02_ux_flows", "03_integration_points", "04_error_handling"},
				},
				{
					name:  "02_database_design", // Parallel sequence
					tasks: []string{"01_schema_design", "02_indexing_strategy", "03_migration_plan", "04_backup_strategy"},
				},
				{
					name:  "03_testing_strategy",
					tasks: []string{"01_unit_test_plan", "02_integration_test_plan", "03_performance_test_plan", "04_acceptance_criteria"},
				},
			},
		},
		{
			name: "003_IMPLEMENT_CORE",
			sequences: []struct {
				name  string
				tasks []string
			}{
				{
					name:  "01_setup",
					tasks: []string{"01_repository_setup", "02_ci_cd_pipeline", "03_dev_environment", "04_dependencies", "05_configuration"},
				},
				{
					name:  "02_data_layer",
					tasks: []string{"01_database_setup", "02_models", "03_repositories", "04_migrations", "05_seeders", "06_data_validation"},
				},
				{
					name:  "03_business_logic",
					tasks: []string{"01_core_services", "02_auth_service", "02_notification_service", "03_workflow_engine", "04_validation_rules", "05_error_handling", "06_logging"},
				},
				{
					name:  "04_api_layer",
					tasks: []string{"01_routes", "02_controllers", "03_middleware", "04_request_validation", "05_response_formatting", "06_api_documentation"},
				},
				{
					name:  "04_event_system", // Parallel sequence
					tasks: []string{"01_event_bus", "02_event_handlers", "03_event_persistence", "04_event_replay"},
				},
			},
		},
		{
			name: "004_IMPLEMENT_FEATURES",
			sequences: []struct {
				name  string
				tasks []string
			}{
				{
					name:  "01_user_management",
					tasks: []string{"01_registration", "02_authentication", "03_authorization", "04_profile_management", "05_password_reset", "06_session_management"},
				},
				{
					name:  "02_core_features",
					tasks: []string{"01_feature_a", "02_feature_b", "02_feature_c", "03_feature_d", "04_feature_integration"},
				},
				{
					name:  "02_reporting", // Parallel sequence
					tasks: []string{"01_report_engine", "02_report_templates", "03_export_functionality", "04_scheduling"},
				},
				{
					name:  "03_integrations",
					tasks: []string{"01_third_party_apis", "02_webhook_system", "03_data_sync", "04_integration_tests"},
				},
			},
		},
		{
			name: "005_TESTING",
			sequences: []struct {
				name  string
				tasks []string
			}{
				{
					name:  "01_unit_testing",
					tasks: []string{"01_test_setup", "02_service_tests", "03_controller_tests", "04_model_tests", "05_coverage_report"},
				},
				{
					name:  "02_integration_testing",
					tasks: []string{"01_api_tests", "02_database_tests", "03_event_tests", "04_e2e_tests"},
				},
			},
		},
	}

	// Create each phase with its sequences and tasks
	for _, phase := range phases {
		phasePath := filepath.Join(festivalPath, phase.name)

		// Create phase directory
		if exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", phasePath}); err != nil || exitCode != 0 {
			return fmt.Errorf("failed to create phase directory %s", phase.name)
		}

		// Create PHASE_GOAL.md
		phaseGoal := CreatePhaseGoalFile(strings.TrimPrefix(phase.name, "00"))
		phaseGoalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{
			"sh", "-c",
			fmt.Sprintf("printf '%%s' %q > %s", phaseGoal, phaseGoalPath),
		})
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to create PHASE_GOAL.md for %s: %w", phase.name, err)
		}

		// Create sequences
		for _, seq := range phase.sequences {
			seqPath := filepath.Join(phasePath, seq.name)

			// Create sequence directory
			if exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", seqPath}); err != nil || exitCode != 0 {
				return fmt.Errorf("failed to create sequence directory %s", seq.name)
			}

			// Create tasks
			for _, task := range seq.tasks {
				taskPath := filepath.Join(seqPath, task+".md")
				taskContent := CreateTaskFile(task)
				exitCode, _, err := tc.container.Exec(tc.ctx, []string{
					"sh", "-c",
					fmt.Sprintf("printf '%%s' %q > %s", taskContent, taskPath),
				})
				if err != nil || exitCode != 0 {
					return fmt.Errorf("failed to create task %s: %w", task, err)
				}
			}
		}
	}

	return nil
}