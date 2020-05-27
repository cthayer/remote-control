package server

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	rc_protocol "github.com/cthayer/go-rc-protocol"
)

var rcProto = rc_protocol.NewRCProtocol()

func TestNewCommand(t *testing.T) {
	msg := rcProto.Message("")
	want := command{
		Stderr:   "",
		Stdout:   "",
		ExitCode: -1,
		Cmd:      "",
		Signal:   nil,
		Timeout:  0,
		Cwd:      "",
		Shell:    []string{"sh", "-c"},
	}

	cmd := newCommand(msg)

	if !reflect.DeepEqual(cmd, want) {
		t.Errorf("newCommand() = %v, wanted %v", cmd, want)
	}
}

func TestCommand_Run(t *testing.T) {
	var msg rc_protocol.Message
	var cmd command
	var timeout int

	// basic command
	msg = rcProto.Message("{\"command\": \"echo 'hello'; echo \\\"world\\\"\"}")

	cmd = newCommand(msg)

	if cmd.Cmd != "echo 'hello'; echo \"world\"" {
		t.Errorf("Command to run not set, wanted: echo 'hello'; echo \"world\", got: %s", cmd.Cmd)
	}

	cmd.Run()

	validateCommandOutput(t, &cmd, "hello\nworld\n", "", 0)

	// command with environment variables
	msg = rcProto.Message("{\"command\": \"echo 'hello'; echo $world\", \"options\": {\"env\": {\"world\": \"mars\"}}}")

	cmd = newCommand(msg)

	if cmd.Cmd != "echo 'hello'; echo $world" {
		t.Errorf("Command to run not set, wanted: echo 'hello'; echo $world, got: %s", cmd.Cmd)
	}

	env := []string{"world=mars"}

	if !reflect.DeepEqual(cmd.Env, env) {
		t.Errorf("Command environment not set, wanted %v, got: %v", env, cmd.Env)
	}

	cmd.Run()

	validateCommandOutput(t, &cmd, "hello\nmars\n", "", 0)

	// command with timeout (runs within timeout)
	timeout = 1000
	msg = rcProto.Message("{\"command\": \"echo 'hello'\", \"options\": {\"timeout\": " + strconv.Itoa(timeout) + "}}")

	cmd = newCommand(msg)

	if cmd.Cmd != "echo 'hello'" {
		t.Errorf("Command to run not set, wanted: echo 'hello', got: %s", cmd.Cmd)
	}

	if cmd.Timeout != timeout {
		t.Errorf("Invalid timeout, wanted %v, got: %v", timeout, cmd.Timeout)
	}

	cmd.Run()

	validateCommandOutput(t, &cmd, "hello\n", "", 0)

	// command with timeout (runs longer than timeout)
	timeout = 100
	msg = rcProto.Message("{\"command\": \"sleep 1; echo 'hello'\", \"options\": {\"timeout\": " + strconv.Itoa(timeout) + "}}")

	cmd = newCommand(msg)

	if cmd.Cmd != "sleep 1; echo 'hello'" {
		t.Errorf("Command to run not set, wanted: sleep 1; echo 'hello', got: %s", cmd.Cmd)
	}

	if cmd.Timeout != timeout {
		t.Errorf("Invalid timeout, wanted %v, got: %v", timeout, cmd.Timeout)
	}

	cmd.Run()

	validateCommandOutput(t, &cmd, "", "", -1)

	// set cwd for command
	pwd, _ := filepath.EvalSymlinks(os.TempDir())
	command := "pwd"
	msg = rcProto.Message("{\"command\": \"" + command + "\", \"options\": {\"cwd\": \"" + pwd + "\"}}")

	cmd = newCommand(msg)

	if cmd.Cmd != command {
		t.Errorf("Command to run not set, wanted: %s, got: %s", command, cmd.Cmd)
	}

	if cmd.Cwd != pwd {
		t.Errorf("Invalid Cwd, wanted %v, got: %v", pwd, cmd.Cwd)
	}

	cmd.Run()

	validateCommandOutput(t, &cmd, pwd+"\n", "", 0)
}

func validateCommandOutput(t *testing.T, cmd *command, expectedStdout string, expectedStderr string, expectedExitCode int) {
	if cmd.ExitCode != expectedExitCode {
		t.Errorf("cmd.Run() = failed, exitcode: %v != %v, stdout: %s, stderr: %s", cmd.ExitCode, expectedExitCode, cmd.Stdout, cmd.Stderr)
	}

	if cmd.Stdout != expectedStdout {
		t.Errorf("stdout: %s != %s", cmd.Stdout, expectedStdout)
	}

	if cmd.Stderr != expectedStderr {
		t.Errorf("stderr: %s != %s", cmd.Stderr, expectedStderr)
	}
}
