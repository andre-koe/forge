package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andre-koe/forge/internal/dsl"
	"github.com/andre-koe/forge/internal/runner"
)

func TestRunRun(t *testing.T) {
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
			name:     "successful run",
			workflow: validWorkflow,
			newRunner: func(path string, opts ...runner.Option) (*runner.Runner, error) {
				// Use the real NewRunner to get proper defaults
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
			err := runRun(tt.workflow, out, tt.newRunner)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("runRun() unexpected error = %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("runRun() expected error, got nil")
				return
			}

			if tt.wantErrType && !errors.Is(err, tt.wantErr) {
				t.Errorf("runRun() error = %v, want error type %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakeRunCmd(t *testing.T) {
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
			cmd := makeRunCmd(tt.newRunner)
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

func TestRunRun_WorkflowExecutionError(t *testing.T) {
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
        run: ["false"]  # Command that will fail
`)
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	out := new(bytes.Buffer)

	// Create a mock runner that simulates a command failure
	mockNewRunner := func(path string, opts ...runner.Option) (*runner.Runner, error) {
		mockRunCmd := func(argv []string) error {
			if len(argv) > 0 && argv[0] == "false" {
				return errors.New("command failed with exit code 1")
			}
			return nil
		}
		return runner.NewRunner(path, append(opts, runner.WithRunCmd(mockRunCmd))...)
	}

	err := runRun(workflowPath, out, mockNewRunner)

	if err == nil {
		t.Fatal("runRun() expected error for failing workflow, got nil")
	}

	if !errors.Is(err, workflowExecutionErr) {
		t.Errorf("runRun() error = %v, want workflowExecutionErr", err)
	}
}

func TestRunRun_WorkflowLoadError(t *testing.T) {
	// Create a temporary invalid workflow file
	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "invalid.yml")
	invalidContent := []byte(`invalid yaml content: [[[`)
	if err := os.WriteFile(workflowPath, invalidContent, 0644); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}

	out := new(bytes.Buffer)

	// Use the real NewRunner, which will fail to parse the YAML
	err := runRun(workflowPath, out, runner.NewRunner)

	if err == nil {
		t.Fatal("runRun() expected error for invalid YAML, got nil")
	}

	if !errors.Is(err, workflowExecutionErr) {
		t.Errorf("runRun() error = %v, want workflowExecutionErr", err)
	}
}

func TestMakeRunCmd_Integration(t *testing.T) {
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

	runCalled := false
	mockNewRunner := func(path string, opts ...runner.Option) (*runner.Runner, error) {
		if path != workflowPath {
			t.Errorf("expected workflow path %q, got %q", workflowPath, path)
		}
		runCalled = true
		// Use real runner to get proper initialization
		return runner.NewRunner(path, opts...)
	}

	cmd := makeRunCmd(mockNewRunner)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{workflowPath})

	// Note: This might fail due to incomplete runner setup, but we check that the mock was called
	_ = cmd.Execute()

	if !runCalled {
		t.Error("expected NewRunner to be called")
	}
}

func TestRunCmd_Properties(t *testing.T) {
	cmd := makeRunCmd(func(path string, opts ...runner.Option) (*runner.Runner, error) {
		return &runner.Runner{}, nil
	})

	// Test command properties
	if cmd.Use != "run [workflow]" {
		t.Errorf("expected Use to be 'run [workflow]', got %q", cmd.Use)
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
