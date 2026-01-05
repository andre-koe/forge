package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWriteTemplate(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		setupFile    bool
		wantErr      error
		wantFileName string
	}{
		{
			name:         "default file name",
			fileName:     "",
			setupFile:    false,
			wantErr:      nil,
			wantFileName: defaultFileName,
		},
		{
			name:         "custom file name",
			fileName:     "custom-workflow.yml",
			setupFile:    false,
			wantErr:      nil,
			wantFileName: "custom-workflow.yml",
		},
		{
			name:         "file already exists",
			fileName:     "existing.yml",
			setupFile:    true,
			wantErr:      errFileExists,
			wantFileName: "existing.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tmpDir := t.TempDir()
			originalWd, _ := os.Getwd()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp dir: %v", err)
			}
			defer func() { _ = os.Chdir(originalWd) }()

			// Setup: create existing file if needed
			if tt.setupFile {
				if err := os.WriteFile(tt.fileName, []byte("existing content"), 0644); err != nil {
					t.Fatalf("failed to setup existing file: %v", err)
				}
			}

			out := new(bytes.Buffer)

			// Execute
			err := runWriteTemplate(tt.fileName, out)

			// Verify error
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("runWriteTemplate() expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("runWriteTemplate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			// Verify no error
			if err != nil {
				t.Errorf("runWriteTemplate() unexpected error = %v", err)
				return
			}

			// Verify output message
			expectedMsg := fmt.Sprintf("Template workflow file created: %s\n", tt.wantFileName)
			if out.String() != expectedMsg {
				t.Errorf("runWriteTemplate() output = %q, want %q", out.String(), expectedMsg)
			}

			// Verify file was created
			if _, err := os.Stat(tt.wantFileName); os.IsNotExist(err) {
				t.Errorf("runWriteTemplate() file %q was not created", tt.wantFileName)
				return
			}

			// Verify file content is not empty
			content, err := os.ReadFile(tt.wantFileName)
			if err != nil {
				t.Fatalf("failed to read created file: %v", err)
			}
			if len(content) == 0 {
				t.Error("runWriteTemplate() created file is empty")
			}

			// Verify file contains expected YAML structure
			expectedKeys := []string{"name:", "description:", "stages:"}
			for _, key := range expectedKeys {
				if !bytes.Contains(content, []byte(key)) {
					t.Errorf("runWriteTemplate() file content missing expected key %q", key)
				}
			}
		})
	}
}

func TestMakeInitCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupFile bool
		wantErr   bool
	}{
		{
			name:      "no arguments - uses default",
			args:      []string{},
			setupFile: false,
			wantErr:   false,
		},
		{
			name:      "custom file name",
			args:      []string{"my-workflow.yml"},
			setupFile: false,
			wantErr:   false,
		},
		{
			name:      "file already exists",
			args:      []string{"existing.yml"},
			setupFile: true,
			wantErr:   true,
		},
		{
			name:      "too many arguments",
			args:      []string{"file1.yml", "file2.yml"},
			setupFile: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tmpDir := t.TempDir()
			originalWd, _ := os.Getwd()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp dir: %v", err)
			}
			defer func() { _ = os.Chdir(originalWd) }()

			// Setup: create existing file if needed
			if tt.setupFile && len(tt.args) > 0 {
				if err := os.WriteFile(tt.args[0], []byte("existing"), 0644); err != nil {
					t.Fatalf("failed to setup existing file: %v", err)
				}
			}

			cmd := makeInitCmd()
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If no error expected, verify file was created
			if !tt.wantErr && err == nil {
				fileName := defaultFileName
				if len(tt.args) > 0 {
					fileName = tt.args[0]
				}
				if _, err := os.Stat(fileName); os.IsNotExist(err) {
					t.Errorf("Execute() file %q was not created", fileName)
				}
			}
		})
	}
}

func TestMakeInitCmd_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	customFileName := "integration-test.yml"

	cmd := makeInitCmd()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{customFileName})

	// First execution should succeed
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("first Execute() unexpected error: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, customFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Execute() file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if len(content) == 0 {
		t.Error("created file is empty")
	}

	// Second execution with same file should fail
	cmd2 := makeInitCmd()
	cmd2.SetOut(out)
	cmd2.SetErr(out)
	cmd2.SetArgs([]string{customFileName})

	err = cmd2.Execute()
	if err == nil {
		t.Error("second Execute() expected error for existing file, got nil")
	}

	if !errors.Is(err, errFileExists) {
		t.Errorf("second Execute() error = %v, want errFileExists", err)
	}
}

func TestInitCmd_Properties(t *testing.T) {
	cmd := makeInitCmd()

	// Test command properties
	if cmd.Use != "init" {
		t.Errorf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be non-empty")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be non-empty")
	}

	// Test that it accepts 0 or 1 argument
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

func TestRunWriteTemplate_EmptyFileName(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// output buffer
	out := new(bytes.Buffer)

	// Empty file name should use default
	err := runWriteTemplate("", out)
	if err != nil {
		t.Errorf("runWriteTemplate(\"\") unexpected error: %v", err)
	}

	// Verify default file was created
	if _, err := os.Stat(defaultFileName); os.IsNotExist(err) {
		t.Errorf("runWriteTemplate(\"\") did not create default file %q", defaultFileName)
	}
}

func TestRunWriteTemplate_PermissionError(t *testing.T) {
	// Skip this test on Windows or if running as root
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	// Create a directory with no write permissions
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	targetFile := filepath.Join(readOnlyDir, "workflow.yml")

	// output buffer
	out := new(bytes.Buffer)

	err := runWriteTemplate(targetFile, out)

	// Should get write failed error
	if err == nil {
		t.Error("runWriteTemplate() expected error for readonly dir, got nil")
	}

	if !errors.Is(err, errWriteFailed) {
		t.Errorf("runWriteTemplate() error = %v, want errWriteFailed", err)
	}
}
