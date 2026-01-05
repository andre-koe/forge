/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/andre-koeniger1997/forge/internal/dsl"
	"github.com/spf13/cobra"
)

const defaultFileName string = "workflow.yaml"

var (
	errFileExists  = errors.New("file already exists")
	errWriteFailed = errors.New("failed to write template file")
)

func runWriteTemplate(fileName string) error {
	if fileName == "" {
		fileName = defaultFileName
	}

	// Check if file already exists
	if _, err := os.Stat(fileName); err == nil {
		return errFileExists
	}

	err := dsl.WriteTemplate(fileName)
	if err != nil {
		return errWriteFailed
	}
	return nil
}

func makeInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Forge project",
		Long: `Initialize a new Forge project by creating a template workflow configuration file.
If a file with the specified name already exists, it will not be overwritten.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runWriteTemplate("")
			}
			return runWriteTemplate(args[0])
		},
	}
}

var initCmd = makeInitCmd()

func init() {
	rootCmd.AddCommand(initCmd)
}
