/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"

	"github.com/andre-koeniger1997/forge/pkg/version"
	"github.com/spf13/cobra"
)

const (
	toolName        = "forge"
	buildDateLabel  = "Build Date"
	commitLabel     = "Git Commit"
	versionCmdUse   = "version"
	versionCmdShort = "Print the version number of Forge"
)

func runVersion(out io.Writer) error {
	fmt.Fprintf(out, "%s %s\n", toolName, version.Version)
	if version.BuildDate != "" {
		fmt.Fprintf(out, "%s: %s\n", buildDateLabel, version.BuildDate)
	}
	if version.Commit != "" {
		fmt.Fprintf(out, "%s: %s\n", commitLabel, version.Commit)
	}
	return nil
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   versionCmdUse,
	Short: versionCmdShort,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion(cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
