package main

import (
	"bufio"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gookit/color"

	rc_protocol "github.com/cthayer/go-rc-protocol"
	"github.com/cthayer/remote_control/internal/logger"
	"github.com/cthayer/remote_control/pkg/client"
	config "github.com/cthayer/remote_control/pkg/client_config"
)

var (
	// #!/usr/bin/env bash
	// version=2
	// time=$(date)
	// go build -ldflags="-X 'main.BuildTime=$time' -X 'main.BuildVersion=$version'" .
	VERSION = "dev"
)

type sendCmdRet struct {
	Host string
	Resp *rc_protocol.Response
	Err  error
}

func main() {
	// setup OS signal handler
	done := setupSignalHandler()

	go func() {
		// run commands
		cmdErr := cliRootCmd.Execute()

		if cmdErr != nil {
			os.Exit(0)
		}

		// signal that execution has completed
		done <- true
	}()

	// wait until execution has completed or an exit signal is received (SIGINT or SIGTERM)
	<-done
}

func runCommand(args []string) {
	// load configuration
	if err := initializeConfig(); err != nil {
		_, _ = os.Stderr.WriteString("Failed to load configuration\n")
		panic(err)
	}

	// setup logger
	log := logger.GetLogger()
	defer log.Sync()

	if len(args) == 1 {
		processStdin(args[0])
		os.Exit(0)
		return
	}

	resp, err := sendCommand(strings.ToLower(strings.TrimSpace(args[0])), args[1], 0)

	writeResponse("", resp, err)

	os.Exit(resp.ExitCode)
}

func processStdin(command string) {
	var batch []string

	batchWaitGroup := sync.WaitGroup{}
	line := bufio.NewScanner(os.Stdin)
	respChan := make(chan sendCmdRet, cliConf.BatchSize)
	firstBatch := true

	for line.Scan() {
		// get the host(s) to send the command to
		host := strings.ToLower(strings.TrimSpace(line.Text()))

		if host == "" {
			continue
		}

		batch = append(batch, host)

		if len(batch) == cliConf.BatchSize {
			// process the batch
			if !firstBatch && cliConf.Delay > 0 {
				// wait between batches
				<-time.After(time.Duration(cliConf.Delay * int(time.Millisecond)))
			} else if firstBatch {
				// if this is the first batch, remember this
				firstBatch = false
			}

			// send the command to all hosts in the batch in parallel
			for _, h := range batch {
				batchWaitGroup.Add(1)
				go handleBackgroundCommand(&batchWaitGroup, h, command, respChan)
			}

			// start a new batch
			batch = []string{}

			// wait for currently executing batch to finish
			batchWaitGroup.Wait()

			// re-sync results to print to terminal properly
			for i := 0; i < cliConf.BatchSize; i++ {
				ret := <-respChan

				writeResponse(ret.Host, ret.Resp, ret.Err)
			}
		}
	}

	if len(batch) > 0 {
		// process any remaining hosts (partial batch size)
		for _, h := range batch {
			batchWaitGroup.Add(1)
			go handleBackgroundCommand(&batchWaitGroup, h, command, respChan)
		}

		// wait for currently executing batch to finish
		batchWaitGroup.Wait()

		// re-sync results to print to terminal properly
		for i := 0; i < cliConf.BatchSize; i++ {
			ret := <-respChan

			writeResponse(ret.Host, ret.Resp, ret.Err)
		}
	}

	if err := line.Err(); err != nil {
		// an error occurred while processing STDIN
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
	}
}

func handleBackgroundCommand(waitGroup *sync.WaitGroup, host string, command string, retChan chan sendCmdRet) {
	defer waitGroup.Done()

	resp, respErr := sendCommand(host, command, 0)

	retChan <- sendCmdRet{
		Host: host,
		Resp: resp,
		Err:  respErr,
	}
}

func writeResponse(host string, resp *rc_protocol.Response, err error) {
	if host != "" {
		_, _ = os.Stdout.WriteString("\n---- " + host + " ----\n")
	}

	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
	}

	if resp == nil {
		_, _ = os.Stderr.WriteString("No response received\n")
		return
	}

	if cliConf.Verbose != 0 {
		// print the raw response object
		jsonStr, err := json.Marshal(resp)

		if err != nil {
			_, _ = os.Stderr.WriteString("Error converting response to json: " + err.Error() + "\n")
			return
		}

		_, _ = os.Stderr.WriteString(string(jsonStr) + "\n")

		return
	}

	red := color.FgRed.Render
	green := color.FgGreen.Render

	_, _ = os.Stderr.WriteString(red(resp.Stderr) + "\n")
	_, _ = os.Stdout.WriteString(green(resp.Stdout) + "\n")
}

func sendCommand(host string, command string, tryCount int) (*rc_protocol.Response, error) {
	log := logger.GetLogger()

	conf := config.Config{
		Port:     cliConf.Port,
		Host:     host,
		KeyDir:   cliConf.KeyDir,
		KeyName:  cliConf.KeyName,
		LogLevel: cliConf.LogLevel,
	}

	conn := client.NewClient(conf)

	errConnect := <-conn.Start()

	if errConnect != nil {
		if tryCount < cliConf.Retry {
			// retry the connection after a slight delay
			log.Debug("connection retry attempt", zap.String("host", host), zap.Int("retry", tryCount+1), zap.Int("maxRetry", cliConf.Retry))
			return sendCommand(host, command, tryCount+1)
		}

		return nil, errConnect
	}

	defer func() {
		_ = <-conn.Stop()
	}()

	resp := <-conn.Send(command, rc_protocol.MessageOptions{})

	return resp, nil
}

func setupSignalHandler() chan bool {
	log := logger.GetLogger()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		defer close(done)

		for {
			sig := <-sigs

			log.Debug("Got signal: " + sig.String())

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				done <- true
			}
		}
	}()

	return done
}
