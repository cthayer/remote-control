package server

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	rc_protocol "github.com/cthayer/go-rc-protocol"
	"github.com/cthayer/remote_control/internal/logger"
)

type command struct {
	Stderr   string
	Stdout   string
	ExitCode int
	Cmd      string
	Signal   os.Signal
	Timeout  int
	Env      []string
	Cwd      string
	Shell    []string
}

func newCommand(msg rc_protocol.Message) command {
	cmd := command{
		Stderr:   "",
		Stdout:   "",
		ExitCode: -1,
		Cmd:      msg.Command,
		Signal:   nil,
		Timeout:  msg.Options.Timeout,
		Cwd:      msg.Options.Cwd,
		Env:      nil,
		Shell:    []string{"sh", "-c"},
	}

	var env []string

	for key, val := range msg.Options.Env {
		env = append(env, key+"="+val)
	}

	cmd.Env = env

	return cmd
}

func (c *command) Run() {
	//if msg.Options.Timeout > 0 {
	//	cmd.Context, cancel = context.WithTimeout(context.Background(), msg.Options.Timeout * time.Millisecond)
	//}
	c.exec()
}

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

	// run the command in it's own process group (needed for graceful shutdowns)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// run the command
	err := cmd.Run()

	// gather the results
	if err != nil {
		log.Debug("Error occurred while running command", zap.Error(err))
		c.ExitCode = cmd.ProcessState.ExitCode()

		if runtime.GOOS != "windows" {
			c.Signal = cmd.ProcessState.Sys().(syscall.WaitStatus).Signal()
		}
	} else {
		c.ExitCode = 0
	}

	c.Stdout = string(stdout.Bytes())
	c.Stderr = string(stderr.Bytes())
}
