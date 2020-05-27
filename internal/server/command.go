package server

import (
	"os"

	rc_protocol "github.com/cthayer/go-rc-protocol"
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
