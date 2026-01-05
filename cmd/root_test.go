package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_Structure(t *testing.T) {
	if rootCmd.Use != "forge" {
		t.Errorf("expected Use to be 'forge', got %q", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("expected Short description to be non-empty")
	}

	// Verify that root command doesn't have a Run function (it's just a container)
	if rootCmd.Run != nil || rootCmd.RunE != nil {
		t.Error("root command should not have Run or RunE defined")
	}
}

func TestRootCmd_SubcommandsRegistered(t *testing.T) {
	expectedSubcommands := []string{"run", "dry-run", "init", "version"}

	for _, name := range expectedSubcommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestRootCmd_HelpOutput(t *testing.T) {
	// Test that help can be displayed without error
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute(--help) unexpected error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}

	// Verify help contains subcommands
	expectedInHelp := []string{"run", "dry-run", "init", "version"}
	for _, cmd := range expectedInHelp {
		if !bytes.Contains([]byte(output), []byte(cmd)) {
			t.Errorf("help output missing subcommand %q", cmd)
		}
	}

	// Reset args for other tests
	rootCmd.SetArgs([]string{})
}

func TestRootCmd_InvalidSubcommand(t *testing.T) {
	out := new(bytes.Buffer)
	cmd := rootCmd
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for nonexistent subcommand, got nil")
	}

	// Reset args
	cmd.SetArgs([]string{})
}

func TestExecute(t *testing.T) {
	// Execute() calls os.Exit on error, so we can't test it directly
	// But we can verify it exists and has the right signature
	// This is more of a compilation check
	_ = Execute
}
