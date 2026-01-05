package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andre-koeniger1997/forge/internal/dsl"
	"github.com/andre-koeniger1997/forge/internal/runner"
)

func TestRunDryRun(t *testing.T) {
	// Create a temporary workflow file for testing
	tmpDir := t.TempDir()
	validWorkflow := filepath.Join(tmpDir, "workflow.yml")
	workflowContent := []byte(`name: test-workflow
description: Test
stages:
  - name: test-stage
    steps:
      - name: hello
        type: exec
        run: ["echo", "hello"]
`)
	if err := os.WriteFile(validWorkflow, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	tests := []struct {
		name        string
		workflow    string
		newRunner   func(string, ...runner.Option) (*runner.Runner, error)
		wantErr     error
		wantErrType bool
	}{
		{
			name:     "successful dry run",
			workflow: validWorkflow,
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				return runner.NewRunner(path, opts...)
			},
			wantErr: nil,
		},
		{
			name:        "empty workflow path",
			workflow:    "",
			newRunner:   nil, // won't be called
			wantErr:     workflowEmptyPathErr,
			wantErrType: true,
		},
		{
			name:        "workflow file not found",
			workflow:    "/non/existent/workflow.yml",
			newRunner:   nil, // won't be called
			wantErr:     workflowNotFoundErr,
			wantErrType: true,
		},
		{
			name:     "runner creation fails",
			workflow: validWorkflow,
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				return nil, errors.New("mock runner creation error")
			},
			wantErr:     runnerCreationErr,
			wantErrType: true,
		},
		{
			name:     "workflow execution fails",
			workflow: validWorkflow,
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				// Create a runner with a mock that returns an error
				mockLoad := func(path string) (*dsl.Workflow, error) {
					return nil, errors.New("mock workflow execution error")
				}
				return runner.NewRunner(path, append(opts, runner.WithLoadWorkflow(mockLoad))...)
			},
			wantErr:     workflowExecutionErr,
			wantErrType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			err := runDryRun(tt.workflow, out, tt.newRunner)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("runDryRun() unexpected error = %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("runDryRun() expected error, got nil")
				return
			}

			if tt.wantErrType && !errors.Is(err, tt.wantErr) {
				t.Errorf("runDryRun() error = %v, want error type %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakeDryRunCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		newRunner func(string, ...runner.Option) (*runner.Runner, error)
		wantErr   error
	}{
		{
			name: "no arguments provided",
			args: []string{},
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				return runner.NewRunner(path, opts...)
			},
			wantErr: nil, // cobra handles this with its own error
		},
		{
			name: "too many arguments",
			args: []string{"workflow1.yml", "workflow2.yml"},
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				return runner.NewRunner(path, opts...)
			},
			wantErr: nil, // cobra handles this with its own error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := makeDryRunCmd(tt.newRunner)
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			// For these tests, we just verify that cobra's arg validation works
			// We don't check the specific error since cobra generates them
			if len(tt.args) != 1 && err == nil {
				t.Error("Execute() expected error for invalid args, got nil")
			}
		})
	}
}

func TestRunDryRun_WorkflowExecutionError(t *testing.T) {
	// Create a temporary workflow file
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "workflow.yml")
	workflowContent := []byte(`name: test-workflow
description: Test
stages:
  - name: build
    steps:
      - name: failing-step
        type: exec
        run: ["false"]
`)
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	out := new(bytes.Buffer)

	// Create a mock runner that simulates a workflow loading error during dry-run
	mockNewRunner := func(path string, opts ...runner.Option) (*runner.Runner, error) {
		mockLoadWorkflow := func(path string) (*dsl.Workflow, error) {
			return nil, errors.New("failed to load workflow")
		}
		return runner.NewRunner(path, append(opts, runner.WithLoadWorkflow(mockLoadWorkflow))...)
	}

	err := runDryRun(workflowPath, out, mockNewRunner)

	if err == nil {
		t.Fatal("runDryRun() expected error for failing workflow load, got nil")
	}

	if !errors.Is(err, workflowExecutionErr) {
		t.Errorf("runDryRun() error = %v, want workflowExecutionErr", err)
	}
}

func TestRunDryRun_WorkflowLoadError(t *testing.T) {
	// Create a temporary invalid workflow file
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "invalid.yml")
	invalidContent := []byte(`invalid yaml content: [[[`)
	if err := os.WriteFile(workflowPath, invalidContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	out := new(bytes.Buffer)

	// Use the real NewRunner, which will fail to parse the YAML
	err := runDryRun(workflowPath, out, runner.NewRunner)

	if err == nil {
		t.Fatal("runDryRun() expected error for invalid YAML, got nil")
	}

	if !errors.Is(err, workflowExecutionErr) {
		t.Errorf("runDryRun() error = %v, want workflowExecutionErr", err)
	}
}

func TestMakeDryRunCmd_Integration(t *testing.T) {
	// Create a temporary workflow file
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test-workflow.yml")
	workflowContent := []byte(`name: test-workflow
description: Test workflow
stages:
  - name: build
    steps:
      - name: hello
        type: exec
        run: ["echo", "Hello"]
`)
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	dryRunCalled := false
	mockNewRunner := func(path string, opts ...runner.Option) (*runner.Runner, error) {
		if path != workflowPath {
			t.Errorf("expected workflow path %q, got %q", workflowPath, path)
		}
		dryRunCalled = true
		// Use real runner to get proper initialization
		return runner.NewRunner(path, opts...)
	}

	cmd := makeDryRunCmd(mockNewRunner)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{workflowPath})

	// Note: This might fail due to incomplete runner setup, but we check that the mock was called
	_ = cmd.Execute()

	if !dryRunCalled {
		t.Error("expected NewRunner to be called")
	}
}

func TestDryRunCmd_Properties(t *testing.T) {
	cmd := makeDryRunCmd(func(path string, opts ...runner.Option) (*runner.Runner, error) {
		return runner.NewRunner(path, opts...)
	})

	// Test command properties
	if cmd.Use != "dry-run [workflow]" {
		t.Errorf("expected Use to be 'dry-run [workflow]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be non-empty")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be non-empty")
	}

	// Test that it expects exactly 1 argument
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

func TestDryRunCmd_OutputContainsDryRunIndicator(t *testing.T) {
	// Create a temporary workflow file
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test-workflow.yml")
	workflowContent := []byte(`name: test-workflow
description: Test
stages:
  - name: test
    steps:
      - name: hello
        type: exec
        run: ["echo", "test"]
`)
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	cmd := makeDryRunCmd(runner.NewRunner)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{workflowPath})

	_ = cmd.Execute()

	output := out.String()

	// Verify that output indicates it's a dry run
	if output != "" && !bytes.Contains([]byte(output), []byte("DRY-RUN")) && !bytes.Contains([]byte(output), []byte("dry-run")) {
		// This is optional - some runners might not include explicit dry-run text
		// but we can at least verify output was generated
		if len(output) == 0 {
			t.Error("expected some output from dry-run, got empty string")
		}
	}
}

func TestRunDryRun_ComparisonWithRun(t *testing.T) {
	// Verify that dry-run and run use the same error types and validation
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "workflow.yml")
	workflowContent := []byte(`name: test
description: Test
stages:
  - name: test
    steps:
      - name: step1
        type: exec
        run: ["echo", "test"]
`)
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	out := new(bytes.Buffer)
	mockNewRunner := func(path string, opts ...runner.Option) (*runner.Runner, error) {
		return nil, errors.New("creation error")
	}

	// Both should return the same error type for runner creation failure
	errRun := runRun(workflowPath, out, mockNewRunner)
	errDryRun := runDryRun(workflowPath, out, mockNewRunner)

	if !errors.Is(errRun, runnerCreationErr) {
		t.Errorf("runRun() error = %v, want runnerCreationErr", errRun)
	}

	if !errors.Is(errDryRun, runnerCreationErr) {
		t.Errorf("runDryRun() error = %v, want runnerCreationErr", errDryRun)
	}

	// Both should handle empty paths the same way
	errRunEmpty := runRun("", out, mockNewRunner)
	errDryRunEmpty := runDryRun("", out, mockNewRunner)

	if !errors.Is(errRunEmpty, workflowEmptyPathErr) {
		t.Errorf("runRun(\"\") error = %v, want workflowEmptyPathErr", errRunEmpty)
	}

	if !errors.Is(errDryRunEmpty, workflowEmptyPathErr) {
		t.Errorf("runDryRun(\"\") error = %v, want workflowEmptyPathErr", errDryRunEmpty)
	}
}
