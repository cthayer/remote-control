package logger

import (
	"go.uber.org/zap"
	"testing"
)

func TestInitLogger(t *testing.T) {
	log, err := InitLogger("info")

	if err != nil {
		t.Errorf("InitLogger() error = %q, wanted %v", err, nil)
	}

	if ent := log.Check(zap.InfoLevel, "foo bar"); ent == nil {
		t.Error("Default loglevel should allow 'info' logs")
	}

	if ent := log.Check(zap.WarnLevel, "foo bar"); ent == nil {
		t.Error("Default loglevel should allow 'warn' logs")
	}

	if ent := log.Check(zap.ErrorLevel, "foo bar"); ent == nil {
		t.Error("Default loglevel should allow 'error' logs")
	}

	if ent := log.Check(zap.DebugLevel, "foo bar"); ent != nil {
		t.Error("Default loglevel should not allow 'debug' logs")
	}
}

func TestGetLogger(t *testing.T) {
	l := GetLogger()

	if l == nil {
		t.Error("Expected *zap.Logger got nil")
	}

	if l != logger {
		t.Errorf("Wanted %v, got %v", logger, l)
	}
}
