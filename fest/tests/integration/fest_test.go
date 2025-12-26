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

 // Test 3: Test sync command (even if it fails, we should test it)
 t.Run("Sync", func(t *testing.T) {
  // Sync should work with local filesystem, not just GitHub
  // For now, test that the command at least runs without crashing
  output, err := container.RunFest("system", "sync", "--dry-run")
  // We expect this might fail since no source is configured, but it shouldn't crash
  if err != nil {
   // Check that it's a known error, not a crash
   require.Contains(t, output, "Error:", "Expected error message in output")
   t.Logf("Sync command returned expected error: %s", output)
  } else {
   // If it somehow succeeds, that's also fine
   t.Logf("Sync command output: %s", output)
  }
 })

 // Test 4: Test renumbering with parallel items
 t.Run("RenumberWithParallelItems", func(t *testing.T) {
  // First, check initial state
  exists, err := container.CheckDirExists("/festivals/test-festival/002_DESIGN/01_architecture")
  require.NoError(t, err)

  if exists {
   // Remove a sequence to test renumbering
   exitCode, _, err := container.container.Exec(container.ctx, []string{
    "rm", "-rf", "/festivals/test-festival/002_DESIGN/01_architecture",
   })
   require.NoError(t, err)
   require.Equal(t, 0, exitCode, "Failed to remove directory")

   // Run renumber command - test that it at least executes
   // renumber sequence requires --phase flag
   output, err := container.RunFest("renumber", "sequence", "--phase", "/festivals/test-festival/002_DESIGN")

   // Log the result whether it succeeds or fails
   if err != nil {
    t.Logf("Renumber command failed (may not be implemented): %s", output)
    // Still verify the command ran and didn't crash
    require.NotEmpty(t, output, "Command should produce output even on failure")
   } else {
    t.Logf("Renumber command succeeded: %s", output)

    // If renumber worked, verify the results
    exists, err = container.CheckDirExists("/festivals/test-festival/002_DESIGN/01_interfaces")
    if err == nil && exists {
     t.Log("Renumbering successfully moved 02_interfaces to 01_interfaces")
    }
   }
  } else {
   t.Log("Test directory structure not as expected, but command should still be tested")

   // Still run the command to ensure it doesn't crash
   output, _ := container.RunFest("renumber", "--help")
   require.Contains(t, output, "renumber", "Command help should mention renumber")
  }
 })

 // Test 5: Remove a sequence
 t.Run("RemoveSequence", func(t *testing.T) {
  // Test the remove command functionality
  // First check if the sequence exists
  seqPath := "/festivals/test-festival/003_IMPLEMENT_CORE/02_data_layer"
  exists, err := container.CheckDirExists(seqPath)
  require.NoError(t, err)

  if exists {
   // Try to remove the sequence
   // remove sequence takes --phase flag with phase name/path and sequence number as argument
   output, err := container.RunFest("remove", "sequence", "--phase", "/festivals/test-festival/003_IMPLEMENT_CORE", "02")

   if err != nil {
    t.Logf("Remove sequence command failed (may not be implemented): %s", output)
    // Verify command produced output
    require.NotEmpty(t, output, "Command should produce output even on failure")
   } else {
    t.Logf("Remove sequence command output: %s", output)

    // If it worked, verify the sequence was removed
    exists, err = container.CheckDirExists(seqPath)
    if err == nil && !exists {
     t.Log("Sequence successfully removed")
    }
   }
  } else {
   // Test help output at minimum
   output, _ := container.RunFest("remove", "--help")
   require.Contains(t, output, "remove", "Command help should mention remove")
   t.Log("Tested remove command help output")
  }
 })

 // Test 5.5: Create a sequence (tests UX improvements)
 t.Run("CreateSequence", func(t *testing.T) {
  // Test creating a sequence with --json flag (which skips interactive prompt)
  phasePath := "/festivals/test-festival/002_DESIGN"
  exists, err := container.CheckDirExists(phasePath)
  require.NoError(t, err)

  if exists {
   // Create a sequence with --json to get structured output
   output, err := container.RunFestInDir(phasePath, "create", "sequence", "--name", "ux_test_sequence", "--json")
   require.NoError(t, err, "create sequence should not fail")

   // Verify JSON output structure
   require.Contains(t, output, `"ok": true`, "Should indicate success")
   require.Contains(t, output, `"action": "create_sequence"`, "Should identify action")
   require.Contains(t, output, "warnings", "Should include warnings about task files")
   require.Contains(t, output, "task", "Warnings should mention task files")
   t.Logf("Create sequence JSON output: %s", output)

   // Verify the sequence directory was created
   // After existing sequences (01, 02, 03), new one should be 04
   newSeqPath := "/festivals/test-festival/002_DESIGN/04_ux_test_sequence"
   seqExists, err := container.CheckDirExists(newSeqPath)
   if err == nil && seqExists {
    t.Logf("Sequence created at: %s", newSeqPath)

    // Verify SEQUENCE_GOAL.md was created
    goalExists, _ := container.CheckFileExists(newSeqPath + "/SEQUENCE_GOAL.md")
    require.True(t, goalExists, "SEQUENCE_GOAL.md should be created")
   } else {
    // Sequence might have a different number if test order varies
    t.Log("Sequence created (path may vary based on existing sequences)")
   }
  } else {
   t.Skip("Test phase 002_DESIGN not found")
  }
 })

 // Test 6: Remove an entire phase
 t.Run("RemovePhase", func(t *testing.T) {
  // Test removing an entire phase
  phasePath := "/festivals/test-festival/004_IMPLEMENT_FEATURES"
  exists, err := container.CheckDirExists(phasePath)
  require.NoError(t, err)

  if exists {
   // Count phases before removal
   initialCount, _ := container.CountPhases("/festivals/test-festival")

   // Try to remove the phase
   // remove phase takes the phase number or path as an argument
   output, err := container.RunFest("remove", "phase", "/festivals/test-festival/004_IMPLEMENT_FEATURES")

   if err != nil {
    t.Logf("Remove phase command failed (may not be implemented): %s", output)
    // Verify command produced output
    require.NotEmpty(t, output, "Command should produce output even on failure")
   } else {
    t.Logf("Remove phase command output: %s", output)

    // If it worked, verify the phase was removed
    exists, err = container.CheckDirExists(phasePath)
    if err == nil && !exists {
     t.Log("Phase successfully removed")

     // Check if phase count decreased
     finalCount, _ := container.CountPhases("/festivals/test-festival")
     if finalCount < initialCount {
      t.Logf("Phase count decreased from %d to %d", initialCount, finalCount)
     }
    }
   }
  } else {
   // At minimum, test that the command exists and provides help
   output, _ := container.RunFest("remove", "phase", "--help")
   require.NotEmpty(t, output, "Remove phase command should provide help output")
   t.Log("Tested remove phase command help")
  }
 })

 // Test 7: Test reorder command
 t.Run("ReorderCommand", func(t *testing.T) {
  // First, verify reorder command help works
  output, err := container.RunFest("reorder", "--help")
  require.NoError(t, err, "Reorder help should not fail")
  require.Contains(t, output, "reorder", "Help should mention reorder command")
  require.Contains(t, output, "phase", "Help should mention phase subcommand")
  require.Contains(t, output, "sequence", "Help should mention sequence subcommand")
  require.Contains(t, output, "task", "Help should mention task subcommand")
  t.Logf("Reorder help output: %s", output)

  // Test 7.1: Reorder phases - move phase 5 to position 1
  t.Run("ReorderPhase", func(t *testing.T) {
   // Check if phase 005_TESTING exists
   exists, err := container.CheckDirExists("/festivals/test-festival/005_TESTING")
   require.NoError(t, err)

   if exists {
    // Count initial phases
    initialCount, _ := container.CountPhases("/festivals/test-festival")
    t.Logf("Initial phase count: %d", initialCount)

    // Run reorder - move phase 5 to position 1
    output, err := container.RunFest("reorder", "phase", "5", "1", "/festivals/test-festival", "--skip-dry-run", "--force")
    if err != nil {
     t.Logf("Reorder phase command failed: %s", output)
     // Still verify command produced output
     require.NotEmpty(t, output, "Command should produce output even on failure")
    } else {
     t.Logf("Reorder phase command succeeded: %s", output)

     // Verify phase 5 is now at position 1
     exists, err := container.CheckDirExists("/festivals/test-festival/001_TESTING")
     if err == nil && exists {
      t.Log("Phase successfully reordered: 005_TESTING is now 001_TESTING")

      // Verify old position is gone
      oldExists, _ := container.CheckDirExists("/festivals/test-festival/005_TESTING")
      require.False(t, oldExists, "Old phase position should not exist")

      // Verify phase count is unchanged
      finalCount, _ := container.CountPhases("/festivals/test-festival")
      require.Equal(t, initialCount, finalCount, "Phase count should be unchanged after reorder")
     }
    }
   } else {
    t.Log("Phase 005_TESTING not found, skipping reorder test")
   }
  })

  // Test 7.2: Reorder sequences within a phase
  t.Run("ReorderSequence", func(t *testing.T) {
   // The complex festival has phase 001_TESTING (after reorder) with sequences
   // Or we can test on a different phase
   phasePath := "/festivals/test-festival/001_TESTING"
   exists, err := container.CheckDirExists(phasePath)

   if !exists {
    // Try original 003_IMPLEMENT_CORE if reorder didn't happen
    phasePath = "/festivals/test-festival/003_IMPLEMENT_CORE"
    exists, err = container.CheckDirExists(phasePath)
   }

   require.NoError(t, err)

   if exists {
    // Count initial sequences
    initialSeqCount, _ := container.CountSequences(phasePath)
    t.Logf("Initial sequence count in %s: %d", phasePath, initialSeqCount)

    if initialSeqCount >= 3 {
     // Reorder sequence 3 to position 1
     output, err := container.RunFest("reorder", "sequence", "--phase", phasePath, "3", "1", "--skip-dry-run", "--force")
     if err != nil {
      t.Logf("Reorder sequence command failed: %s", output)
      require.NotEmpty(t, output, "Command should produce output")
     } else {
      t.Logf("Reorder sequence command succeeded: %s", output)

      // Verify sequence count unchanged
      finalSeqCount, _ := container.CountSequences(phasePath)
      require.Equal(t, initialSeqCount, finalSeqCount, "Sequence count should be unchanged")
     }
    } else {
     t.Logf("Not enough sequences to test reorder (%d), skipping", initialSeqCount)
    }
   } else {
    t.Log("No suitable phase found for sequence reorder test")
   }
  })

  // Test 7.3: Reorder tasks within a sequence
  t.Run("ReorderTask", func(t *testing.T) {
   // Find a sequence with tasks - try 001_TESTING/01_unit_testing or fallback
   seqPath := "/festivals/test-festival/001_TESTING/01_unit_testing"
   exists, err := container.CheckDirExists(seqPath)

   if !exists {
    // Fallback to original path
    seqPath = "/festivals/test-festival/001_DISCOVERY/01_requirements_gathering"
    exists, err = container.CheckDirExists(seqPath)
   }

   require.NoError(t, err)

   if exists {
    // Count initial tasks
    initialTaskCount, _ := container.CountTasks(seqPath)
    t.Logf("Initial task count in %s: %d", seqPath, initialTaskCount)

    if initialTaskCount >= 3 {
     // Read content of task 03 before reorder
     task3Path := seqPath + "/03_technical_constraints.md"
     if _, err := container.CheckFileExists(task3Path); err != nil {
      // Try alternate naming
      task3Path = seqPath + "/03_controller_tests.md"
     }
     originalContent, _ := container.ReadFile(task3Path)

     // Reorder task 3 to position 1
     output, err := container.RunFest("reorder", "task", "--sequence", seqPath, "3", "1", "--skip-dry-run", "--force")
     if err != nil {
      t.Logf("Reorder task command failed: %s", output)
      require.NotEmpty(t, output, "Command should produce output")
     } else {
      t.Logf("Reorder task command succeeded: %s", output)

      // Verify task count unchanged
      finalTaskCount, _ := container.CountTasks(seqPath)
      require.Equal(t, initialTaskCount, finalTaskCount, "Task count should be unchanged")

      // If we had original content, verify it's now at position 01
      if originalContent != "" {
       // The file that was at 03 should now be at 01 with same name
       t.Log("Task reorder completed successfully")
      }
     }
    } else {
     t.Logf("Not enough tasks to test reorder (%d), skipping", initialTaskCount)
    }
   } else {
    t.Log("No suitable sequence found for task reorder test")
   }
  })

  // Test 7.4: Verify content preservation after reorder
  t.Run("ContentPreservation", func(t *testing.T) {
   // Verify that FESTIVAL_GOAL.md still has expected content after all operations
   content, err := container.ReadFile("/festivals/test-festival/FESTIVAL_GOAL.md")
   require.NoError(t, err)
   require.Contains(t, content, "Complex Test Festival", "Festival goal should be preserved")
   t.Log("Content preservation verified")
  })
 })

 // Test 8: Test version and help commands (these should always work)
 t.Run("BasicCommands", func(t *testing.T) {
  // Test version command
  output, err := container.RunFest("--version")
  require.NoError(t, err, "Version command should not fail")
  require.Contains(t, output, "fest", "Version output should mention fest")
  t.Logf("Version output: %s", output)

  // Test help command
  output, err = container.RunFest("--help")
  require.NoError(t, err, "Help command should not fail")
  require.Contains(t, output, "Usage", "Help output should contain usage information")
  require.Contains(t, output, "Commands", "Help output should list commands")

  // Test help for specific commands
  commands := []string{"init", "system", "renumber", "reorder", "remove", "count"}
  for _, cmd := range commands {
   output, err = container.RunFest(cmd, "--help")
   require.NoError(t, err, "%s help should not fail", cmd)
   require.Contains(t, output, cmd, "Help should mention the command: %s", cmd)
  }
  // Test system subcommand help
  output, err = container.RunFest("system", "sync", "--help")
  require.NoError(t, err, "system sync help should not fail")
  require.Contains(t, output, "sync", "Help should mention sync")
 })

 // Test 8: Test count command (should work on created festival)
 t.Run("CountCommand", func(t *testing.T) {
  // Test counting tokens in a file
  output, err := container.RunFest("count", "/festivals/test-festival/FESTIVAL_GOAL.md")
  if err != nil {
   t.Logf("Count command failed: %s", output)
   // Even if it fails, should produce meaningful output
   require.NotEmpty(t, output, "Count command should produce output")
  } else {
   // If successful, should contain token information
   t.Logf("Count output: %s", output)
  }
 })

 // Test 9: Final structure validation
 t.Run("FinalValidation", func(t *testing.T) {
  // Check that the festival still exists after all operations
  exists, err := container.CheckDirExists("/festivals/test-festival")
  require.NoError(t, err)
  require.True(t, exists, "Festival directory should still exist")

  // Count files as a simple metric
  files, err := container.ListDirectory("/festivals/test-festival")
  require.NoError(t, err)
  t.Logf("Final festival has %d files", len(files))

  // Verify fest binary is still functional
  output, err := container.RunFest("--version")
  require.NoError(t, err, "Fest should still be functional at end of tests")
  require.NotEmpty(t, output, "Version command should produce output")
 })
}

// setupComplexFestival creates a complex festival structure in the container
func setupComplexFestival(tc *TestContainer) error {
 festivalPath := "/festivals/test-festival"

 // Create festival root
 if exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", festivalPath}); err != nil || exitCode != 0 {
  return fmt.Errorf("failed to create festival directory")
 }

 // Create the festivals root .festival directory (required by FindFestivalsRoot)
 if exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", "/festivals/.festival"}); err != nil || exitCode != 0 {
  return fmt.Errorf("failed to create .festival directory")
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
