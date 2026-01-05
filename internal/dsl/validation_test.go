package dsl

import "testing"

func TestValidateSteps(t *testing.T) {
	tests := []struct {
		name    string
		step    Step
		wantErr bool
	}{
		{
			name: "valid exec step",
			step: Step{
				Name: "step1",
				Type: StepTypeExec,
				Run:  []string{"echo", "Hello"},
			},
			wantErr: false,
		},
		{
			name: "valid sleep step",
			step: Step{
				Name:    "step2",
				Type:    StepTypeSleep,
				Seconds: 5,
			},
			wantErr: false,
		},
		{
			name: "missing step name",
			step: Step{
				Type: StepTypeExec,
				Run:  []string{"echo", "Hello"},
			},
			wantErr: true,
		},
		{
			name: "missing step type",
			step: Step{
				Name: "step3",
				Run:  []string{"echo", "Hello"},
			},
			wantErr: true,
		},
		{
			name: "exec step missing run",
			step: Step{
				Name: "step4",
				Type: StepTypeExec,
			},
			wantErr: true,
		},
		{
			name: "sleep step with non-positive seconds",
			step: Step{
				Name:    "step5",
				Type:    StepTypeSleep,
				Seconds: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.step.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStages(t *testing.T) {
	tests := []struct {
		name    string
		stage   Stage
		wantErr bool
	}{
		{
			name: "valid stage",
			stage: Stage{
				Name: "stage1",
				Steps: []Step{
					{
						Name: "step1",
						Type: StepTypeExec,
						Run:  []string{"echo", "Hello"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing stage name",
			stage: Stage{
				Steps: []Step{
					{
						Name: "step1",
						Type: StepTypeExec,
						Run:  []string{"echo", "Hello"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "stage with no steps",
			stage: Stage{
				Name:  "stage2",
				Steps: []Step{},
			},
			wantErr: true,
		},
		{
			name: "stage with invalid step",
			stage: Stage{
				Name: "stage3",
				Steps: []Step{
					{
						Name: "step2",
						Type: StepTypeExec,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stage.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWorkflows(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		wantErr  bool
	}{
		{
			name: "valid workflow",
			workflow: Workflow{
				Name: "workflow1",
				Stages: []Stage{
					{
						Name: "stage1",
						Steps: []Step{
							{
								Name: "step1",
								Type: StepTypeExec,
								Run:  []string{"echo", "Hello"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing workflow name",
			workflow: Workflow{
				Stages: []Stage{
					{
						Name: "stage1",
						Steps: []Step{
							{
								Name: "step1",
								Type: StepTypeExec,
								Run:  []string{"echo", "Hello"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "workflow with no stages",
			workflow: Workflow{
				Name:   "workflow2",
				Stages: []Stage{},
			},
			wantErr: true,
		},
		{
			name: "workflow with invalid stage",
			workflow: Workflow{
				Name: "workflow3",
				Stages: []Stage{
					{
						Name:  "stage2",
						Steps: []Step{},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
