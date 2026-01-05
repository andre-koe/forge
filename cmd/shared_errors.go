package cmd

import (
	"errors"
	"os"
)

var (
	workflowEmptyPathErr = errors.New("workflow path cannot be empty")
	workflowNotFoundErr  = errors.New("workflow file not found")
	runnerCreationErr    = errors.New("failed to create runner")
	workflowExecutionErr = errors.New("workflow execution failed")
)

func CheckFilePathExistAndIsNotEmpty(path string) error {
	if path == "" {
		return workflowEmptyPathErr
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return workflowNotFoundErr
	}
	return nil
}
