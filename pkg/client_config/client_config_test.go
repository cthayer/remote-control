package client_config

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	want := Config{
		Port:     DEFAULT_PORT,
		Host:     DEFAULT_HOST,
		KeyDir:   DEFAULT_KEY_DIR,
		KeyName:  DEFAULT_KEY_NAME,
		LogLevel: DEFAULT_LOG_LEVEL,
	}

	if config := GetConfig(); !reflect.DeepEqual(*config, want) {
		t.Errorf("GetConfig() = %v, wanted %v", config, want)
	}
}
