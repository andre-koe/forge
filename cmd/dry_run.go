/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"

	"github.com/andre-koe/forge/internal/runner"
	"github.com/spf13/cobra"
)

func runDryRun(workflow string, out io.Writer, newRunner func(string, ...runner.Option) (*runner.Runner, error)) error {
	if err := CheckFilePathExistAndIsNotEmpty(workflow); err != nil {
		return err
	}

	r, err := newRunner(workflow, runner.WithOut(out))
	if err != nil {
		return runnerCreationErr
	}

	if err := r.DryRun(); err != nil {
		return workflowExecutionErr
	}
	return nil
}

func makeDryRunCmd(newRunner func(string, ...runner.Option) (*runner.Runner, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "dry-run [workflow]",
		Short: "Simulate the execution of a workflow without making any changes",
		Long:  `Simulate the execution of a workflow defined in your forge configuration file without making any changes.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDryRun(args[0], cmd.OutOrStdout(), newRunner)
		},
	}
}

var (
	dryRunCmd = makeDryRunCmd(runner.NewRunner)
)

func init() {
	rootCmd.AddCommand(dryRunCmd)
}
