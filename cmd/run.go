package cmd

import (
	"io"

	"github.com/andre-koeniger1997/forge/internal/runner"
	"github.com/spf13/cobra"
)

func runRun(workflow string, out io.Writer, newRunner func(string, ...runner.Option) (*runner.Runner, error)) error {
	if err := CheckFilePathExistAndIsNotEmpty(workflow); err != nil {
		return err
	}

	r, err := newRunner(workflow, runner.WithOut(out))
	if err != nil {
		return runnerCreationErr
	}

	if err := r.Run(); err != nil {
		return workflowExecutionErr
	}

	return nil
}

func makeRunCmd(newRunner func(string, ...runner.Option) (*runner.Runner, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "run [workflow]",
		Short: "Execute a defined workflow",
		Long:  `Execute a workflow defined in your forge configuration file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(args[0], cmd.OutOrStdout(), newRunner)
		},
	}
}

var runCmd = makeRunCmd(runner.NewRunner)

func init() {
	rootCmd.AddCommand(runCmd)
}
