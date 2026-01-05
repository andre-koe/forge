package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/andre-koe/forge/internal/dsl"
)

// Options for configuring the Runner

type Option func(*Runner)

func WithOut(w io.Writer) Option {
	return func(r *Runner) {
		r.Out = w
	}
}

func WithLoadWorkflow(f func(path string) (*dsl.Workflow, error)) Option {
	return func(r *Runner) {
		r.LoadWorkflow = f
	}
}

func WithRunCmd(f func(argv []string) error) Option {
	return func(r *Runner) {
		r.RunCmd = f
	}
}

func WithSleep(f func(d time.Duration)) Option {
	return func(r *Runner) { r.Sleep = f }
}

// Runner implements Runner
type Runner struct {
	path         string
	LoadWorkflow func(path string) (*dsl.Workflow, error)
	RunCmd       func(argv []string) error
	Sleep        func(d time.Duration)
	Out          io.Writer
}

// NewRunner creates a new Runner for the specified workflow
func NewRunner(path string, opts ...Option) (*Runner, error) {
	r := &Runner{
		path:         path,
		LoadWorkflow: dsl.LoadWorkflowFromFile,
		RunCmd:       runCommand,
		Sleep:        time.Sleep,
		Out:          os.Stdout,
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.LoadWorkflow == nil || r.RunCmd == nil || r.Sleep == nil || r.Out == nil {
		return nil, fmt.Errorf("runner not properly configured")
	}
	return r, nil
}

func (r *Runner) Run() error {
	fmt.Fprintf(r.Out, "Executing workflow: %s\n", r.path)

	wf, err := r.LoadWorkflow(r.path)
	if err != nil {
		return err
	}

	// Iterate through stages
	// TODO: Allow for parallel stage and or step execution in the future
	for stageIdx, stage := range wf.Stages {
		fmt.Fprintf(r.Out, "\n=== STAGE %d: %s ===\n", stageIdx+1, stage.Name)

		// Execute each step in the stage
		for stepIdx, step := range stage.Steps {
			fmt.Fprintf(r.Out, "STEP %d.%d: %s (%s)\n", stageIdx+1, stepIdx+1, step.Name, step.Type)

			if err := r.executeStep(&step); err != nil {
				return fmt.Errorf("stage '%s', step '%s': %w", stage.Name, step.Name, err)
			}
		}

		fmt.Fprintf(r.Out, "=== STAGE %d COMPLETED ===\n", stageIdx+1)
	}

	fmt.Fprintf(r.Out, "\n✓ Workflow execution completed.\n")
	return nil
}

// DryRun simulates Workflow execution
func (r *Runner) DryRun() error {
	fmt.Fprintf(r.Out, "[DRY-RUN] Would execute workflow: %s\n", r.path)

	wf, err := r.LoadWorkflow(r.path)
	if err != nil {
		return err
	}

	// Iterate through stages
	// TODO: Allow for parallel stage and or step "simulation" in the future
	for stageIdx, stage := range wf.Stages {
		fmt.Fprintf(r.Out, "\n[DRY-RUN] === STAGE %d: %s ===\n", stageIdx+1, stage.Name)

		// Simulate each step in the stage
		for stepIdx, step := range stage.Steps {
			fmt.Fprintf(r.Out, "[DRY-RUN] STEP %d.%d: %s (%s)\n", stageIdx+1, stepIdx+1, step.Name, step.Type)

			switch step.Type {
			case dsl.StepTypeExec:
				fmt.Fprintf(r.Out, "[DRY-RUN]   Would execute command: %v\n", step.Run)
			case dsl.StepTypeSleep:
				fmt.Fprintf(r.Out, "[DRY-RUN]   Would sleep for %d seconds\n", step.Seconds)
			}
		}

		fmt.Fprintf(r.Out, "[DRY-RUN] === STAGE %d COMPLETED ===\n", stageIdx+1)
	}

	fmt.Fprintf(r.Out, "\n[DRY-RUN] ✓ Workflow simulation completed.\n")
	return nil
}

// executeStep executes a single step (extracted for reusability)
func (r *Runner) executeStep(step *dsl.Step) error {
	switch step.Type {
	case dsl.StepTypeExec:
		if err := r.RunCmd(step.Run); err != nil {
			return fmt.Errorf("command execution failed: %w", err)
		}
	case dsl.StepTypeSleep:
		fmt.Fprintf(r.Out, "  Sleeping for %d seconds...\n", step.Seconds)
		r.Sleep(time.Duration(step.Seconds) * time.Second)
	default:
		// This should never happen due to validation in LoadWorkflowFromFile
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
	return nil
}

// runCommand executes a command with arguments
func runCommand(argv []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
