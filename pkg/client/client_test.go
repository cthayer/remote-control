package client

import (
	"path/filepath"
	"testing"

	rc_protocol "github.com/cthayer/go-rc-protocol"
	server_config "github.com/cthayer/remote_control/internal/config"
	"github.com/cthayer/remote_control/internal/server"
	config "github.com/cthayer/remote_control/pkg/client_config"
)

func TestNewClient(t *testing.T) {
	conf := config.GetConfig()

	client := NewClient(*conf)

	if _, ok := client.(Client); !ok {
		t.Errorf("Invalid interface type. client is not of type 'Client'")
	}
}

func TestClient_Start_Stop_No_Auth(t *testing.T) {
	conf := config.GetConfig()

	client := NewClient(*conf)

	srv, _ := startServer(t)

	defer stopServer(t, srv)

	// should fail to connect when not using the proper authentication
	errChan := client.Start()

	err := <-errChan

	if err.Error() != "websocket: bad handshake" {
		t.Error("Client did not fail authentication when it should")
	}

	// stop should be safe to call even when there is no connection
	errChan2 := client.Stop()

	err2 := <-errChan2

	if err2 != nil {
		t.Errorf("Error stopping client: %v", err2)
	}
}

func TestClient_Start_Stop_Auth(t *testing.T) {
	conf := config.GetConfig()

	conf.KeyName = "client"
	conf.KeyDir = filepath.Join("..", "..", "test", "client", "keys")

	client := NewClient(*conf)

	srv, _ := startServer(t)

	defer stopServer(t, srv)

	// should connect when using the proper authentication
	errChan := client.Start()

	err := <-errChan

	if err != nil {
		t.Errorf("Error connecting with proper authentication: %v", err)
	}

	// stop should be safe to call even when there is no connection
	errChan2 := client.Stop()

	err2 := <-errChan2

	if err2 != nil {
		t.Errorf("Error stopping client: %v", err2)
	}
}

func TestClient_Send(t *testing.T) {
	srv, _ := startServer(t)
	defer stopServer(t, srv)

	client, _ := startClient(t)
	defer stopClient(t, client)

	respChan := (*client).Send("echo hello world", rc_protocol.MessageOptions{})

	resp := <-respChan

	validateResponse(t, resp, "hello world\n", "", 0)
}

func TestClient_Send_Invalid_Command(t *testing.T) {
	srv, _ := startServer(t)
	defer stopServer(t, srv)

	client, _ := startClient(t)
	defer stopClient(t, client)

	respChan := (*client).Send("foo hello world", rc_protocol.MessageOptions{})

	resp := <-respChan

	validateResponse(t, resp, "", "sh: foo: command not found\n", 127)
}

func startClient(t *testing.T) (*Client, error) {
	conf := config.GetConfig()

	conf.KeyName = "client"
	conf.KeyDir = filepath.Join("..", "..", "test", "client", "keys")

	client := NewClient(*conf)

	errChan := client.Start()

	err := <-errChan

	if err != nil {
		t.Errorf("Error staring client: %v", err)
	}

	return &client, err
}

func stopClient(t *testing.T, client *Client) {
	errChan := (*client).Stop()

	err := <-errChan

	if err != nil {
		t.Errorf("Error stopping client: %v", err)
	}
}

func startServer(t *testing.T) (*server.Server, error) {
	conf := server_config.GetConfig()

	conf.CertDir = filepath.Join("..", "..", "test", "server", "certs")

	server := server.NewServer(conf)

	errChan := server.Start()

	err := <-errChan

	if err != nil {
		t.Errorf("Error starting server: %v", err)
	}

	return &server, err
}

func stopServer(t *testing.T, server *server.Server) {
	errChan := (*server).Stop()

	err := <-errChan

	if err != nil {
		t.Errorf("Error stopping server: %v", err)
	}
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
