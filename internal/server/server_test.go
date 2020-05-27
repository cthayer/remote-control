package server

import (
	"github.com/cthayer/remote_control/pkg/client"
	"github.com/cthayer/remote_control/pkg/client_config"
	"path/filepath"
	"testing"
	"time"

	"github.com/cthayer/remote_control/internal/config"

	rc_protocol "github.com/cthayer/go-rc-protocol"
)

func TestNewServer(t *testing.T) {
	conf := config.GetConfig()

	server := NewServer(conf)

	if _, ok := server.(Server); !ok {
		t.Errorf("Invalid interface type. server is not of type 'Server'")
	}
}

func TestServer_Start_Stop(t *testing.T) {
	conf := config.GetConfig()

	server := NewServer(conf)

	errChan := server.Start()

	err := <-errChan

	if err != nil {
		t.Errorf("Error starting server: %v", err)
		return
	}

	errChan2 := server.Stop()

	err2 := <-errChan2

	if err2 != nil {
		t.Errorf("Error stopping server: %v", err2)
	}
}

func TestServer_Send_Command(t *testing.T) {
	server, err := startServer(t)

	if err != nil {
		return
	}

	defer stopServer(t, server)

	// send a command to the server
	resp := sendMessage(t, "echo 'hello'; echo \"world\"", rc_protocol.MessageOptions{})

	if resp == nil {
		t.Errorf("No response received: %v", resp)
	}

	validateResponse(t, resp, "hello\nworld\n", "", 0)
}

func startServer(t *testing.T) (*Server, error) {
	conf := config.GetConfig()

	conf.CertDir = filepath.Join("..", "..", "test", "server", "certs")

	server := NewServer(conf)

	errChan := server.Start()

	err := <-errChan

	if err != nil {
		t.Errorf("Error starting server: %v", err)
	}

	return &server, err
}

func stopServer(t *testing.T, server *Server) {
	errChan := (*server).Stop()

	err := <-errChan

	if err != nil {
		t.Errorf("Error stopping server: %v", err)
	}
}

func sendMessage(t *testing.T, command string, msgOptions rc_protocol.MessageOptions) *rc_protocol.Response {
	var resp *rc_protocol.Response = nil

	clientConf := client_config.GetConfig()
	clientConf.KeyDir = filepath.Join("..", "..", "test", "client", "keys")
	clientConf.KeyName = "client"

	client := client.NewClient(*clientConf)
	defer client.Stop()

	startErrChan := client.Start()

	select {
	case startErr := <-startErrChan:
		if startErr != nil {
			t.Errorf("Error while connecting to server: %v", startErr)
			return resp
		}
	case <-time.After(time.Second):
		t.Error("Timeout exceeded while connecting to server")
		return resp
	}

	respChan := client.Send(command, msgOptions)

	resp = <-respChan

	return resp
}

func validateResponse(t *testing.T, cmd *rc_protocol.Response, expectedStdout string, expectedStderr string, expectedExitCode int) {
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
