package logger

import (
	"encoding/json"
	"go.uber.org/zap"
)

var logger *zap.Logger

func InitLogger(logLevel string) (*zap.Logger, error) {
	rawJSON := []byte(`{
	  "level": "` + logLevel + `",
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	var err error

	if err = json.Unmarshal(rawJSON, &cfg); err != nil {
		return nil, err
	}

	logger, err = cfg.Build()

	return logger, err
}

func GetLogger() *zap.Logger {
	return logger
}
