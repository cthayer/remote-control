package client_config

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	want := Config{
		Port:          DEFAULT_PORT,
		Host:          DEFAULT_HOST,
		KeyDir:        DEFAULT_KEY_DIR,
		KeyName:       DEFAULT_KEY_NAME,
		LogLevel:      DEFAULT_LOG_LEVEL,
		TlsSkipVerify: DEFAULT_TLS_SKIP_VERIFY,
		TlsCaFile:     DEFAULT_TLS_CA_FILE,
		TlsDisable:    DEFAULT_TLS_DISABLE,
	}

	if config := GetConfig(); !reflect.DeepEqual(*config, want) {
		t.Errorf("GetConfig() = %v, wanted %v", config, want)
	}
}
