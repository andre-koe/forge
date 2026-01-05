package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andre-koe/forge/pkg/version"
)

func TestRunVersion(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		buildDate      string
		commit         string
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:           "version only",
			version:        "1.0.0",
			buildDate:      "",
			commit:         "",
			wantContain:    []string{"forge 1.0.0"},
			wantNotContain: []string{"Build Date:", "Git Commit:"},
		},
		{
			name:           "version and build Date",
			version:        "1.0.0",
			buildDate:      "2026-01-01",
			commit:         "",
			wantContain:    []string{"forge 1.0.0", "Build Date: 2026-01-01"},
			wantNotContain: []string{"Git Commit:"},
		},
		{
			name:           "version and commit",
			version:        "1.0.0",
			buildDate:      "",
			commit:         "abc123",
			wantContain:    []string{"forge 1.0.0", "Git Commit: abc123"},
			wantNotContain: []string{"Build Date:"},
		},
		{
			name:        "all fields",
			version:     "1.0.0",
			buildDate:   "2024-01-01",
			commit:      "abc123",
			wantContain: []string{"forge 1.0.0", "Build Date: 2024-01-01", "Git Commit: abc123"},
		},
		{
			name:           "dev version",
			version:        "dev",
			buildDate:      "",
			commit:         "",
			wantContain:    []string{"forge dev"},
			wantNotContain: []string{"Build Date:", "Git Commit:"},
		},
		{
			name:      "version with special characters",
			version:   "v2.1.0-beta.1+sha.5114f85",
			buildDate: "2024-12-01T10:30:00Z",
			commit:    "5114f85abc",
			wantContain: []string{
				"forge v2.1.0-beta.1+sha.5114f85",
				"Build Date: 2024-12-01T10:30:00Z",
				"Git Commit: 5114f85abc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup and restore original values
			origVersion := version.Version
			origBuildDate := version.BuildDate
			origCommit := version.Commit
			t.Cleanup(func() {
				version.Version = origVersion
				version.BuildDate = origBuildDate
				version.Commit = origCommit
			})

			// Set test values
			version.Version = tt.version
			version.BuildDate = tt.buildDate
			version.Commit = tt.commit

			out := new(bytes.Buffer)
			err := runVersion(out)
			if err != nil {
				t.Fatalf("runVersion() unexpected error: %v", err)
			}

			output := out.String()

			// Check for expected content
			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected content:\ngot:  %q\nwant: %q", output, want)
				}
			}

			// Check for unexpected content
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("output contains unexpected content:\ngot:  %q\nshould not contain: %q", output, notWant)
				}
			}
		})
	}
}

func TestVersionCmd(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		buildDate      string
		commit         string
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:      "version command with all fields",
			version:   "1.2.3",
			buildDate: "2024-01-15",
			commit:    "abc123def",
			wantContain: []string{
				"forge 1.2.3",
				"Build Date: 2024-01-15",
				"Git Commit: abc123def",
			},
		},
		{
			name:           "version command with minimal info",
			version:        "0.1.0",
			buildDate:      "",
			commit:         "",
			wantContain:    []string{"forge 0.1.0"},
			wantNotContain: []string{"Build Date:", "Git Commit:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup and restore original values
			origVersion := version.Version
			origBuildDate := version.BuildDate
			origCommit := version.Commit
			t.Cleanup(func() {
				version.Version = origVersion
				version.BuildDate = origBuildDate
				version.Commit = origCommit
			})

			// Set test values
			version.Version = tt.version
			version.BuildDate = tt.buildDate
			version.Commit = tt.commit

			// Capture stdout
			out := new(bytes.Buffer)
			versionCmd.SetOut(out)
			versionCmd.SetErr(out)

			// Execute command
			err := versionCmd.RunE(versionCmd, []string{})
			if err != nil {
				t.Fatalf("versionCmd.RunE() unexpected error: %v", err)
			}

			output := out.String()

			// Check for expected content
			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected content:\ngot:  %q\nwant: %q", output, want)
				}
			}

			// Check for unexpected content
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("output contains unexpected content:\ngot:  %q\nshould not contain: %q", output, notWant)
				}
			}
		})
	}
}

func TestVersionCmdProperties(t *testing.T) {
	tests := []struct {
		name     string
		property string
		want     string
	}{
		{
			name:     "command use",
			property: "Use",
			want:     versionCmdUse,
		},
		{
			name:     "command short description",
			property: "Short",
			want:     versionCmdShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.property {
			case "Use":
				got = versionCmd.Use
			case "Short":
				got = versionCmd.Short
			}

			if got != tt.want {
				t.Errorf("versionCmd.%s = %q, want %q", tt.property, got, tt.want)
			}
		})
	}
}

func TestVersionCmdRunEReturnsNil(t *testing.T) {
	// Backup original values
	origVersion := version.Version
	origBuildDate := version.BuildDate
	origCommit := version.Commit
	t.Cleanup(func() {
		version.Version = origVersion
		version.BuildDate = origBuildDate
		version.Commit = origCommit
	})

	version.Version = "test"
	version.BuildDate = ""
	version.Commit = ""

	out := new(bytes.Buffer)
	versionCmd.SetOut(out)

	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Errorf("versionCmd.RunE() should return nil, got error: %v", err)
	}
}
