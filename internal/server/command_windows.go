//+build windows

package server

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"go.uber.org/zap"

	"github.com/cthayer/remote_control/internal/logger"
)

func (c *command) exec() {
	var cmd *exec.Cmd

	log := logger.GetLogger()

	fullCmd := append(c.Shell, c.Cmd)

	// set a timeout for the command (if specified)
	if c.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Millisecond)
		defer cancel()

		cmd = exec.CommandContext(ctx, fullCmd[0], fullCmd[1:]...)
	} else {
		cmd = exec.Command(fullCmd[0], fullCmd[1:]...)
	}

	// set the environment and working directory for the command
	cmd.Dir = c.Cwd
	cmd.Env = c.Env

	// setup capturing of stdout and stderr
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run the command
	err := cmd.Run()

	// gather the results
	if err != nil {
		log.Debug("Error occurred while running command", zap.Error(err))
		c.ExitCode = cmd.ProcessState.ExitCode()
	} else {
		c.ExitCode = 0
	}

	c.Stdout = string(stdout.Bytes())
	c.Stderr = string(stderr.Bytes())
}
