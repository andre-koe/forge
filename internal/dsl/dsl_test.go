package dsl

import (
	"os"
	"path/filepath"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

func TestWriteTemplate(t *testing.T) {
	tests := []struct {
		name       string
		setupPath  func(t *testing.T) string
		wantErr    bool
		validateFn func(t *testing.T, path string)
	}{
		{
			name: "successful template creation",
			setupPath: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "template_workflow.yaml")
			},
			wantErr: false,
			validateFn: func(t *testing.T, path string) {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("ReadFile(%s) error: %v", path, err)
				}

				var wf Workflow
				if err := yaml.Unmarshal(data, &wf); err != nil {
					t.Fatalf("Unmarshal() error: %v", err)
				}

				if wf.Name != "example-forge-workflow" {
					t.Errorf("wf.Name = %q, want %q", wf.Name, "example-forge-workflow")
				}
				if len(wf.Stages) != 2 {
					t.Fatalf("len(wf.Stages) = %d, want %d", len(wf.Stages), 2)
				}

				firstStage := wf.Stages[0]
				if firstStage.Name != "hello-stage" {
					t.Errorf("firstStage.Name = %q, want %q", firstStage.Name, "hello-stage")
				}
				if len(firstStage.Steps) != 2 {
					t.Fatalf("len(firstStage.Steps) = %d, want %d", len(firstStage.Steps), 2)
				}

				firstStep := firstStage.Steps[0]
				if firstStep.Name != "hello" || firstStep.Type != StepTypeExec {
					t.Errorf("firstStep mismatch: got Name=%q Type=%q, want Name=%q Type=%q",
						firstStep.Name, firstStep.Type, "hello", StepTypeExec)
				}

				secondStage := wf.Stages[1]
				if secondStage.Name != "goodbye-stage" {
					t.Errorf("secondStage.Name = %q, want %q", secondStage.Name, "goodbye-stage")
				}
				if len(secondStage.Steps) != 1 {
					t.Fatalf("len(secondStage.Steps) = %d, want %d", len(secondStage.Steps), 1)
				}

				secondStep := secondStage.Steps[0]
				if secondStep.Name != "goodbye" || secondStep.Type != StepTypeExec {
					t.Errorf("secondStep mismatch: got Name=%q Type=%q, want Name=%q Type=%q",
						secondStep.Name, secondStep.Type, "goodbye", StepTypeExec)
				}
			},
		},
		{
			name: "path is directory returns error",
			setupPath: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath(t)
			err := WriteTemplate(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFn != nil {
				tt.validateFn(t, path)
			}
		})
	}
}

func TestLoadWorkflowFromFile(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(t *testing.T) string
		wantErr    bool
		validateFn func(t *testing.T, wf *Workflow)
	}{
		{
			name: "valid workflow file",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "sample_workflow.yaml")
				content := `
name: Sample Workflow
description: A sample workflow for testing
stages:
  - name: stage1
    steps:
      - name: step1
        type: exec
        run:
          - echo
          - Hello World
      - name: step2
        type: sleep
        seconds: 2
`
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: false,
			validateFn: func(t *testing.T, wf *Workflow) {
				if wf.Name != "Sample Workflow" {
					t.Errorf("wf.Name = %q, want %q", wf.Name, "Sample Workflow")
				}
				if len(wf.Stages) != 1 {
					t.Fatalf("len(wf.Stages) = %d, want %d", len(wf.Stages), 1)
				}

				firstStage := wf.Stages[0]
				if firstStage.Name != "stage1" {
					t.Errorf("firstStage.Name = %q, want %q", firstStage.Name, "stage1")
				}
				if len(firstStage.Steps) != 2 {
					t.Fatalf("len(firstStage.Steps) = %d, want %d", len(firstStage.Steps), 2)
				}

				first := firstStage.Steps[0]
				if first.Name != "step1" || first.Type != StepTypeExec {
					t.Errorf("first step mismatch: got Name=%q Type=%q, want Name=%q Type=%q",
						first.Name, first.Type, "step1", StepTypeExec)
				}

				second := firstStage.Steps[1]
				if second.Name != "step2" || second.Type != StepTypeSleep {
					t.Errorf("second step mismatch: got Name=%q Type=%q, want Name=%q Type=%q",
						second.Name, second.Type, "step2", StepTypeSleep)
				}
				if second.Seconds != 2 {
					t.Errorf("second.Seconds = %d, want %d", second.Seconds, 2)
				}
			},
		},
		{
			name: "non-existent file",
			setupFile: func(t *testing.T) string {
				return "non_existent_file.yaml"
			},
			wantErr: true,
		},
		{
			name: "invalid YAML",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "invalid_workflow.yaml")
				content := `
name: Invalid Workflow
description: "This workflow has invalid YAML"
steps:
  - name: step1
    type: exec
    run: ["echo", "Hello World"
  - name: step2
    type: sleep
    seconds: 2
`
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: true,
		},
		{
			name: "validation error - missing workflow name",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "invalid_workflow.yaml")
				content := `
description: A workflow without a name
stages:
  - name: stage1
    steps:
    - name: step1
      type: exec
        run:
          - echo
          - Hello World
`
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: true,
		},
		{
			name: "validation error - stage with no steps",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "invalid_workflow.yaml")
				content := `
name: Workflow With Invalid Stage
description: A workflow with a stage that has no steps
stages:
  - name: empty-stage
    steps: []
`
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: true,
		},
		{
			name: "validation error - exec step with missing run command",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "invalid_workflow.yaml")
				content := `
name: Workflow With Invalid Step
description: A workflow with an exec step missing the run command
stages:
  - name: stage1
    steps:
    - name: invalid-step
      type: exec
`
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: true,
		},
		{
			name: "validation success - should handle workflow with tabs",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				filename := filepath.Join(dir, "workflow_with_tabs.yaml")
				content := "name: Workflow With Tabs\n" +
					"description: A workflow file that uses tabs for indentation\n" +
					"stages:\n" +
					"\t- name: tabbed-stage\n" +
					"\t  steps:\n" +
					"\t\t- name: tabbed-step\n" +
					"\t\t  type: exec\n" +
					"\t\t  run:\n" +
					"\t\t\t- echo\n" +
					"\t\t\t- Hello from tabs\n"
				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s) error: %v", filename, err)
				}
				return filename
			},
			wantErr: false,
			validateFn: func(t *testing.T, wf *Workflow) {
				if wf.Name != "Workflow With Tabs" {
					t.Errorf("wf.Name = %q, want %q", wf.Name, "Workflow With Tabs")
				}
				if len(wf.Stages) != 1 {
					t.Fatalf("len(wf.Stages) = %d, want %d", len(wf.Stages), 1)
				}

				stage := wf.Stages[0]
				if stage.Name != "tabbed-stage" {
					t.Errorf("stage.Name = %q, want %q", stage.Name, "tabbed-stage")
				}
				if len(stage.Steps) != 1 {
					t.Fatalf("len(stage.Steps) = %d, want %d", len(stage.Steps), 1)
				}

				step := stage.Steps[0]
				if step.Name != "tabbed-step" || step.Type != StepTypeExec {
					t.Errorf("step mismatch: got Name=%q Type=%q, want Name=%q Type=%q",
						step.Name, step.Type, "tabbed-step", StepTypeExec)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.setupFile(t)
			wf, err := LoadWorkflowFromFile(filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadWorkflowFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFn != nil {
				tt.validateFn(t, wf)
			}
		})
	}
}
