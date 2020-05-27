package client

import (
	"encoding/json"
	"golang.org/x/tools/container/intsets"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cthayer/remote_control/internal/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	rc_protocol "github.com/cthayer/go-rc-protocol"
	config "github.com/cthayer/remote_control/pkg/client_config"
)

const (
	CLOSE_TIMEOUT    = 120
	WEBSOCKET_SCHEME = "ws"
	WEBSOCKET_PATH   = "/"
)

var rcProto rc_protocol.RCProtocol = rc_protocol.NewRCProtocol()

type Client interface {
	Start() chan error
	Stop() chan error
	Send(string, rc_protocol.MessageOptions) chan *rc_protocol.Response
}

type client struct {
	socket       *websocket.Conn
	conf         config.Config
	logger       *zap.Logger
	isConnected  bool
	url          url.URL
	readLoopDone chan struct{}
	msgChannels  map[int]chan rc_protocol.Response
	msgId        int
}

func NewClient(conf config.Config) Client {
	c := client{
		conf:         conf,
		logger:       logger.GetLogger(),
		isConnected:  false,
		url:          url.URL{Scheme: WEBSOCKET_SCHEME, Host: conf.Host + ":" + strconv.Itoa(conf.Port), Path: WEBSOCKET_PATH},
		readLoopDone: nil,
		msgChannels:  map[int]chan rc_protocol.Response{},
		msgId:        0,
	}

	return &c
}

func (c *client) Start() chan error {
	errChan := make(chan error, 1)

	go func() {
		var err error = nil

		defer close(errChan)

		if c.isConnected {
			errChan <- err
			return
		}

		// create authorization header
		reqHeader := c.createSig()

		c.socket, _, err = websocket.DefaultDialer.Dial(c.url.String(), reqHeader)

		if err != nil {
			c.logger.Error("Failed to connect to server.", zap.String("url", c.url.String()), zap.Error(err))
		}

		c.isConnected = err == nil

		if c.isConnected {
			c.readLoopDone = make(chan struct{})

			// start read loop in the background
			go c.readMessages()
		}

		errChan <- err
	}()

	return errChan
}

func (c *client) Stop() chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		var err error = nil

		if !c.isConnected {
			errChan <- err
			return
		}

		// Attempt to cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err = c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

		if err != nil {
			c.logger.Error("Error sending close message", zap.String("url", c.url.String()), zap.Any("socket", c.socket), zap.Error(err))

			// hard close the connection
			err2 := c.socket.Close()

			if err2 != nil {
				c.logger.Error("Error closing connection to server", zap.String("url", c.url.String()), zap.Any("socket", c.socket), zap.Error(err2))
			}

			// reset the socket regardless of an error during disconnect
			c.resetConn()

			errChan <- err

			return
		}

		select {
		case <-c.readLoopDone:
		case <-time.After(CLOSE_TIMEOUT * time.Second):
			// hard close the connection
			err = c.socket.Close()
		}

		if err != nil {
			c.logger.Error("Error closing connection to server.", zap.String("url", c.url.String()), zap.Error(err))
		}

		// reset the socket regardless of an error during disconnect
		c.resetConn()

		errChan <- err
	}()

	return errChan
}

func (c *client) Send(command string, options rc_protocol.MessageOptions) chan *rc_protocol.Response {
	respChan := make(chan *rc_protocol.Response, 1)
	var resp *rc_protocol.Response = nil

	msg := rc_protocol.Message{
		Command: command,
		Id:      c.nextMessageId(),
		Options: options,
	}

	c.msgChannels[msg.Id] = make(chan rc_protocol.Response, 1)

	go func() {
		defer func() {
			// close the message channels
			close(respChan)
			close(c.msgChannels[msg.Id])

			// remove the message channel from the channel map
			delete(c.msgChannels, msg.Id)
		}()

		// must be connected to send the message
		if !c.isConnected {
			respChan <- resp
			return
		}

		// prep message to send to the server
		jsonStr, jErr := json.Marshal(msg)

		if jErr != nil {
			c.logger.Error("Error converting msg to json", zap.String("url", c.url.String()), zap.Any("msg", msg), zap.Any("socket", c.socket), zap.Error(jErr))

			respChan <- resp

			return
		}

		// send the message to the server
		err := c.socket.WriteMessage(websocket.TextMessage, jsonStr)

		if err != nil {
			c.logger.Error("Error writing message to socket", zap.String("url", c.url.String()), zap.Any("socket", c.socket), zap.ByteString("json", jsonStr), zap.Error(err))

			respChan <- resp

			return
		}

		// wait for the response
		resp := <-c.msgChannels[msg.Id]

		// send response to caller
		respChan <- &resp
	}()

	return respChan
}

func (c *client) createSig() http.Header {
	header := http.Header{}

	sig, err := rcProto.CreateSig(c.conf.KeyName, c.conf.KeyDir)

	if err != nil {
		c.logger.Error("Error creating signature", zap.Error(err))
		return header
	}

	header.Add("Authorization", sig)

	return header
}

func (c *client) nextMessageId() int {
	if c.msgId >= intsets.MaxInt-1 {
		// wrap around to 0 if we exhaust the range of an integer
		c.msgId = 0
	}

	// increment the message id
	c.msgId = c.msgId + 1

	return c.msgId
}

func (c *client) readMessages() {
	defer close(c.readLoopDone)

	for {
		messageType, message, err := c.socket.ReadMessage()

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				c.logger.Debug("Connection closed normally", zap.String("url", c.url.String()), zap.Any("socket", c.socket), zap.Error(err))
			} else {
				c.logger.Error("Error reading message", zap.String("url", c.url.String()), zap.Any("socket", c.socket), zap.Error(err))
			}

			return
		}

		c.logger.Debug("websocket message received", zap.Any("message", message))

		if messageType != websocket.TextMessage {
			// this is not the response to the request
			continue
		}

		// convert the raw message to an rc_protocol.Response object
		resp := rcProto.Response(string(message))

		c.logger.Debug("response received", zap.Any("resp", resp))

		msgId, errId := strconv.Atoi(resp.Id)

		if errId != nil {
			c.logger.Error("Error converting response id to integer", zap.String("response.id", resp.Id), zap.Error(errId), zap.Any("response", resp))

			continue
		}

		if msgChan, ok := c.msgChannels[msgId]; ok {
			msgChan <- resp
		}
	}
}

func (c *client) resetConn() {
	c.socket = nil
	c.isConnected = false
	c.readLoopDone = nil
	c.msgChannels = map[int]chan rc_protocol.Response{}
	c.msgId = 0
}
