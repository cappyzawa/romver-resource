package resource

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Runnner runs command
type Runner interface {
	Run(cmd *exec.Cmd) error
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	Error() error
}

type ExCommand struct {
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

func (ec *ExCommand) Run(cmd *exec.Cmd) error {
	cmd.Stdout = ec.Stdout
	cmd.Stderr = ec.Stderr
	return cmd.Run()
}

func (ec *ExCommand) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (ec *ExCommand) Error() error {
	if ec.Stderr != nil {
		return fmt.Errorf(ec.Stderr.String())
	}
	return nil
}
