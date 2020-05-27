package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/cthayer/remote_control/internal/config"
	"github.com/cthayer/remote_control/internal/logger"
	"github.com/cthayer/remote_control/internal/server"
)

const (
	SERVER_START_TIMEOUT = 120
)

func main() {
	// load configuration
	if err := initializeConfig(); err != nil {
		_, _ = os.Stderr.WriteString("Failed to load configuration\n")
		panic(err)
	}

	log := logger.GetLogger()
	defer log.Sync()

	log.Info("Configuration Loaded")

	// start server
	srv := server.NewServer(config.GetConfig())

	errChan := srv.Start()

	select {
	case startErr := <-errChan:
		if startErr != nil {
			log.Error("Server failed to start", zap.Error(startErr))
			log.Info("Exiting")
			return
		}
	case <-time.After(SERVER_START_TIMEOUT * time.Second):
		log.Error("Giving up on starting the server.")
		log.Info("Exiting")
		return
	}

	// setup OS signal handler
	done := setupSignalHandler()

	// wait until signaled to exit (SIGINT or SIGTERM) -- SIGHUP to reload config
	<-done

	log.Info("Shutting down")

	// stop the server
	errChan = srv.Stop()

	select {
	case stopErr := <-errChan:
		if stopErr != nil {
			log.Error("Server failed to stop", zap.Error(stopErr))
			log.Info("Exiting")
			return
		}
	case <-done: // if another OS interrupt is received, exit hard
		log.Error("Giving up on stopping the server.")
		log.Info("Exiting")
		return
	}

	log.Info("Shutdown complete")
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
			case syscall.SIGHUP:
				reloadConfig()
			}
		}
	}()

	return done
}
