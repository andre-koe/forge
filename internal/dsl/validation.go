package dsl

import (
	"errors"
	"fmt"
)

func (w *Workflow) Validate() error {
	if w.Name == "" {
		return errors.New("workflow name is required")
	}

	if len(w.Stages) == 0 {
		return errors.New("workflow must have at least one stage")
	}

	for i, stage := range w.Stages {
		if err := stage.Validate(); err != nil {
			return fmt.Errorf("stage %d (%s): %w", i, stage.Name, err)
		}
	}

	return nil
}

// Validate validates a stage
func (s *Stage) Validate() error {
	if s.Name == "" {
		return errors.New("stage name is required")
	}

	if len(s.Steps) == 0 {
		return errors.New("stage must have at least one step")
	}

	for i, step := range s.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d (%s): %w", i, step.Name, err)
		}
	}

	return nil
}

// Validate validates a step
func (s *Step) Validate() error {
	if s.Name == "" {
		return errors.New("step name is required")
	}

	switch s.Type {
	case StepTypeExec:
		if len(s.Run) == 0 {
			return errors.New("exec step requires 'run' command")
		}
	case StepTypeSleep:
		if s.Seconds <= 0 {
			return errors.New("sleep step requires positive 'seconds' value")
		}
	default:
		return fmt.Errorf("unknown step type: %s", s.Type)
	}

	return nil
}
