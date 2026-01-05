package runner

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andre-koe/forge/internal/dsl"
)

func mockLoadWorkflow(stages []dsl.Stage) func(path string) (*dsl.Workflow, error) {
	// Mock workflow for testing
	return func(path string) (*dsl.Workflow, error) {
		wf := &dsl.Workflow{
			Name:   "mock-workflow",
			Stages: stages,
		}
		return wf, nil
	}
}

func mockLoadWorkflowError(err error) func(path string) (*dsl.Workflow, error) {
	return func(path string) (*dsl.Workflow, error) {
		return nil, err
	}
}

func mockRunCmd(calls *[][]string) func(argv []string) error {
	return func(argv []string) error {
		*calls = append(*calls, argv)
		return nil
	}
}

func mockRunCmdError(err error) func(argv []string) error {
	return func(argv []string) error {
		return err
	}
}

func mockSleep(calls *[]time.Duration) func(d time.Duration) {
	return func(d time.Duration) {
		*calls = append(*calls, d)
	}
}

func TestNewRunner(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "test.yaml",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "with custom out",
			path: "test.yaml",
			opts: []Option{
				WithOut(new(bytes.Buffer)),
			},
			wantErr: false,
		},
		{
			name: "with nil out should fail",
			path: "test.yaml",
			opts: []Option{
				WithOut(nil),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.path, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRunner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && runner == nil {
				t.Error("NewRunner() returned nil runner")
			}
		})
	}
}

func TestRunner_Run_Success(t *testing.T) {
	var cmdCalls [][]string
	var sleepCalls []time.Duration
	out := new(bytes.Buffer)

	workflow := []dsl.Stage{
		{
			Name: "build",
			Steps: []dsl.Step{
				{
					Name: "compile",
					Type: dsl.StepTypeExec,
					Run:  []string{"go", "build"},
				},
				{
					Name:    "wait",
					Type:    dsl.StepTypeSleep,
					Seconds: 2,
				},
			},
		},
		{
			Name: "test",
			Steps: []dsl.Step{
				{
					Name: "run-tests",
					Type: dsl.StepTypeExec,
					Run:  []string{"go", "test"},
				},
			},
		},
	}

	runner, err := NewRunner("test-workflow.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflow(workflow)),
		WithRunCmd(mockRunCmd(&cmdCalls)),
		WithSleep(mockSleep(&sleepCalls)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify command calls
	if len(cmdCalls) != 2 {
		t.Errorf("expected 2 command calls, got %d", len(cmdCalls))
	}
	if !slices.Equal(cmdCalls[0], []string{"go", "build"}) {
		t.Errorf("expected first command ['go', 'build'], got %v", cmdCalls[0])
	}
	if !slices.Equal(cmdCalls[1], []string{"go", "test"}) {
		t.Errorf("expected second command ['go', 'test'], got %v", cmdCalls[1])
	}

	// Verify sleep calls
	if len(sleepCalls) != 1 {
		t.Errorf("expected 1 sleep call, got %d", len(sleepCalls))
	}
	if sleepCalls[0] != 2*time.Second {
		t.Errorf("expected sleep duration 2s, got %v", sleepCalls[0])
	}

	// Verify output
	output := out.String()
	if !strings.Contains(output, "STAGE 1: build") {
		t.Error("output missing stage 1 header")
	}
	if !strings.Contains(output, "STAGE 2: test") {
		t.Error("output missing stage 2 header")
	}
	if !strings.Contains(output, "STEP 1.1: compile (exec)") {
		t.Error("output missing step 1.1")
	}
	if !strings.Contains(output, "✓ Workflow execution completed") {
		t.Error("output missing completion message")
	}
}

func TestRunner_Run_LoadWorkflowError(t *testing.T) {
	out := new(bytes.Buffer)
	expectedErr := errors.New("failed to load workflow")

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflowError(expectedErr)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.Run()
	if err == nil {
		t.Fatal("Run() should have failed")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRunner_Run_CommandError(t *testing.T) {
	out := new(bytes.Buffer)
	expectedErr := errors.New("command failed")

	workflow := []dsl.Stage{
		{
			Name: "build",
			Steps: []dsl.Step{
				{
					Name: "compile",
					Type: dsl.StepTypeExec,
					Run:  []string{"go", "build"},
				},
			},
		},
	}

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflow(workflow)),
		WithRunCmd(mockRunCmdError(expectedErr)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.Run()
	if err == nil {
		t.Fatal("Run() should have failed")
	}
	if !strings.Contains(err.Error(), "stage 'build'") {
		t.Errorf("error should mention stage name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "step 'compile'") {
		t.Errorf("error should mention step name, got: %v", err)
	}
}

func TestRunner_DryRun_Success(t *testing.T) {
	out := new(bytes.Buffer)

	workflow := []dsl.Stage{
		{
			Name: "deploy",
			Steps: []dsl.Step{
				{
					Name: "terraform-apply",
					Type: dsl.StepTypeExec,
					Run:  []string{"terraform", "apply"},
				},
				{
					Name:    "wait-for-deployment",
					Type:    dsl.StepTypeSleep,
					Seconds: 10,
				},
			},
		},
	}

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflow(workflow)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.DryRun()
	if err != nil {
		t.Fatalf("DryRun() failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Error("output missing [DRY-RUN] prefix")
	}
	if !strings.Contains(output, "STAGE 1: deploy") {
		t.Error("output missing stage header")
	}
	if !strings.Contains(output, "Would execute command: [terraform apply]") {
		t.Error("output missing command simulation")
	}
	if !strings.Contains(output, "Would sleep for 10 seconds") {
		t.Error("output missing sleep simulation")
	}
	if !strings.Contains(output, "✓ Workflow simulation completed") {
		t.Error("output missing completion message")
	}
}

func TestRunner_DryRun_LoadWorkflowError(t *testing.T) {
	out := new(bytes.Buffer)
	expectedErr := errors.New("failed to load workflow")

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflowError(expectedErr)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.DryRun()
	if err == nil {
		t.Fatal("DryRun() should have failed")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRunner_ExecuteStep_UnknownType(t *testing.T) {
	out := new(bytes.Buffer)

	runner := &Runner{
		Out: out,
	}

	step := &dsl.Step{
		Name: "unknown-step",
		Type: dsl.StepType("unknown"),
	}

	err := runner.executeStep(step)
	if err == nil {
		t.Fatal("executeStep() should have failed for unknown step type")
	}
	if !strings.Contains(err.Error(), "unknown step type") {
		t.Errorf("error should mention unknown step type, got: %v", err)
	}
}

func TestRunner_MultipleStages(t *testing.T) {
	var cmdCalls [][]string
	out := new(bytes.Buffer)

	workflow := []dsl.Stage{
		{
			Name: "stage1",
			Steps: []dsl.Step{
				{Name: "step1", Type: dsl.StepTypeExec, Run: []string{"cmd1"}},
			},
		},
		{
			Name: "stage2",
			Steps: []dsl.Step{
				{Name: "step2", Type: dsl.StepTypeExec, Run: []string{"cmd2"}},
			},
		},
		{
			Name: "stage3",
			Steps: []dsl.Step{
				{Name: "step3", Type: dsl.StepTypeExec, Run: []string{"cmd3"}},
			},
		},
	}

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflow(workflow)),
		WithRunCmd(mockRunCmd(&cmdCalls)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if len(cmdCalls) != 3 {
		t.Errorf("expected 3 command calls, got %d", len(cmdCalls))
	}

	output := out.String()
	for i := 1; i <= 3; i++ {
		stageHeader := "STAGE " + string(rune('0'+i))
		if !strings.Contains(output, stageHeader) {
			t.Errorf("output missing stage %d header", i)
		}
	}
}

func TestRunner_EmptyWorkflow(t *testing.T) {
	out := new(bytes.Buffer)

	workflow := []dsl.Stage{}

	runner, err := NewRunner("test.yaml",
		WithOut(out),
		WithLoadWorkflow(mockLoadWorkflow(workflow)),
	)
	if err != nil {
		t.Fatalf("NewRunner() failed: %v", err)
	}

	err = runner.Run()
	if err != nil {
		t.Fatalf("Run() should handle empty workflow, got error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✓ Workflow execution completed") {
		t.Error("output missing completion message for empty workflow")
	}
}
