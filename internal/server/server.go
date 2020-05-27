package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	rc_protocol "github.com/cthayer/go-rc-protocol"
	"github.com/cthayer/remote_control/internal/config"
	"github.com/cthayer/remote_control/internal/logger"
)

const (
	COMMAND_QUEUE_MAX_BACKLOG = 5
	MAX_CONCURRENT_COMMANDS   = 5
	HTTP_SERVER_STOP_TIMEOUT  = 300 //5 minutes are allowed to stop the http server
)

type Server interface {
	Start() chan error
	Stop() chan error
}

type server struct {
	conf               config.Config
	upgrader           websocket.Upgrader
	logger             *zap.Logger
	rcProto            rc_protocol.RCProtocol
	cmdQueue           chan commandQueue
	httpSrv            *http.Server
	netListener        net.Listener
	router             *mux.Router
	waitGroup          sync.WaitGroup
	shutdown           chan struct{}
	cmdWorkerWaitGroup sync.WaitGroup
}

type commandQueue struct {
	Command  command             `json:"command"`
	Message  rc_protocol.Message `json:"message"`
	RespChan chan commandResp    `json:"-"`
}

type commandResp struct {
	Response rc_protocol.Response
	Error    error
}

func NewServer(conf *config.Config) Server {
	srv := server{
		conf: *conf,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger:             logger.GetLogger(),
		rcProto:            rc_protocol.NewRCProtocol(),
		cmdQueue:           make(chan commandQueue, COMMAND_QUEUE_MAX_BACKLOG),
		httpSrv:            &http.Server{Addr: conf.Host + ":" + strconv.Itoa(conf.Port)},
		netListener:        nil,
		router:             mux.NewRouter(),
		waitGroup:          sync.WaitGroup{},
		shutdown:           make(chan struct{}),
		cmdWorkerWaitGroup: sync.WaitGroup{},
	}

	return &srv
}

func (s *server) Start() chan error {
	var err error = nil

	errChan := make(chan error, 1)
	s.shutdown = make(chan struct{})

	// start the server async
	s.waitGroup.Add(1)
	go func() {
		// start the command processing loop
		s.waitGroup.Add(1)
		go s.runCommands()

		s.logger.Debug("Command processing go routine started")

		s.router.HandleFunc("/", s.handler)
		s.httpSrv.Handler = s.router

		s.netListener, err = net.Listen("tcp", s.httpSrv.Addr)

		errChan <- err
		close(errChan)

		if err != nil {
			return
		}

		s.logger.Info("Server started", zap.String("listen address", s.httpSrv.Addr))

		// this will block until the server is stopped or an error occurs
		err = s.httpSrv.Serve(s.netListener)

		// if the server exits for any reason other than being stopped, log the error
		if err != http.ErrServerClosed {
			s.logger.Error("Error serving requests", zap.Error(err))
		}

		s.waitGroup.Done()
	}()

	return errChan
}

func (s *server) Stop() chan error {
	defer close(s.shutdown)

	errChan := make(chan error, 1)

	// stop the server async
	go func() {
		defer close(errChan)

		// stop server
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*HTTP_SERVER_STOP_TIMEOUT)
		err := s.httpSrv.Shutdown(ctx)

		cancel()

		s.logger.Debug("HTTP server shutdown")

		s.logger.Debug("Waiting for commands to finish running")

		// wait for shutdown to finish
		s.waitGroup.Wait()

		errChan <- err
	}()

	return errChan
}

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	// check authorization header
	authHeader := r.Header.Get(s.rcProto.GetHeaderName())

	validSig, err := s.rcProto.CheckSig(authHeader, s.conf.CertDir)

	if err != nil {
		s.logger.Error("Error occurred while checking signature", zap.Error(err))
		return
	}

	if !validSig {
		s.logger.Error("Invalid signature", zap.Bool("validSig", validSig))
		return
	}

	s.logger.Debug("Client authenticated successfully")

	// upgrade request to websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		s.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}

	s.logger.Debug("Succeeded in upgrading to websocket")

	//conn.SetCloseHandler(func(code int, text string) error {
	//	s.logger.Info("connection close handler fired", zap.Int("code", code), zap.String("text", text))
	//	message := websocket.FormatCloseMessage(code, "")
	//	_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
	//	return nil
	//})

	// handle websocket messages
	s.waitGroup.Add(1)
	go s.websocketHandler(conn)
}

func (s *server) websocketHandler(conn *websocket.Conn) {
	defer s.closeConn(conn)
	defer s.waitGroup.Done()

commLoop:
	for {
		messageType, p, err := conn.ReadMessage()

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				s.logger.Debug("Normal websocket close", zap.Error(err), zap.Any("conn", conn))
			} else {
				s.logger.Error("Error reading message from websocket", zap.Error(err), zap.Any("conn", conn))
			}

			break commLoop
		}

		s.logger.Debug("received message from client", zap.Int("messageType", messageType), zap.ByteString("message", p))

		switch messageType {
		case websocket.BinaryMessage:
			// invalid type
			s.logger.Error("Binary Messages are not accepted")
			break commLoop
		case websocket.TextMessage:
			resp, err := s.handleMessage(string(p))

			if err != nil {
				s.logger.Error("Error handling message", zap.Error(err), zap.Any("conn", conn))
				break commLoop
			}

			s.logger.Debug("succeeded in handling message", zap.Any("response", *resp))

			jsonResp, err := json.Marshal(resp)

			if err != nil {
				s.logger.Error("Error marshalling json response", zap.Error(err), zap.Any("resp", resp), zap.Any("conn", conn))
				break commLoop
			}

			s.logger.Debug("converted message to json", zap.ByteString("json", jsonResp))

			if err := conn.WriteMessage(websocket.TextMessage, jsonResp); err != nil {
				s.logger.Error("Error writing message to socket", zap.Error(err), zap.Any("conn", conn))
				break commLoop
			}

			s.logger.Debug("Sent message to client")
		}
	}
}

func (s *server) closeConn(conn *websocket.Conn) {
	s.logger.Debug("Closing connection", zap.Any("conn", conn))

	err := conn.Close()

	if err != nil {
		s.logger.Error("Error while closing connection", zap.Error(err), zap.Any("conn", conn))
	}
}

func (s *server) handleMessage(msg string) (*rc_protocol.Response, error) {
	message := s.rcProto.Message(msg)
	respChan := make(chan commandResp, 1)

	cmd := commandQueue{
		Command:  newCommand(message),
		Message:  message,
		RespChan: respChan,
	}

	select {
	case s.cmdQueue <- cmd:
		s.logger.Debug("command added to queue", zap.Any("command", cmd))
	case <-time.After(time.Millisecond):
		return nil, errors.New("command queue is full.  " + string(COMMAND_QUEUE_MAX_BACKLOG) + " commands waiting to run")
	}

	resp := <-respChan

	return &resp.Response, resp.Error
}

func (s *server) runCommands() {
	defer s.waitGroup.Done()

	for i := 0; i < MAX_CONCURRENT_COMMANDS; i++ {
		s.cmdWorkerWaitGroup.Add(1)
		go s.commandWorker(i)
	}

	// wait for workers to finish
	s.cmdWorkerWaitGroup.Wait()
}

func (s *server) commandWorker(workerId int) {
	defer s.cmdWorkerWaitGroup.Done()

	for {
		var c commandQueue

		select {
		case c = <-s.cmdQueue:
			s.logger.Debug("Running command", zap.Any("command", c), zap.Int("workerId", workerId))
		case <-s.shutdown:
			s.logger.Debug("Server shutting down.  Breaking command loop", zap.Int("workerId", workerId))
			return
		}

		c.Command.Run()

		jsonResp, err := json.Marshal(c.Command)

		resp := commandResp{
			Response: s.rcProto.Response(string(jsonResp)),
			Error:    err,
		}

		resp.Response.Id = strconv.Itoa(c.Message.Id)

		c.RespChan <- resp
		close(c.RespChan)

		s.logger.Debug("Finished running command", zap.Any("command", c), zap.Int("workerId", workerId))
	}
}
