package config

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	want := Config{
		Port:        DEFAULT_PORT,
		Host:        DEFAULT_HOST,
		CertDir:     DEFAULT_CERT_DIR,
		Ciphers:     DEFAULT_CIPHERS,
		LogLevel:    DEFAULT_LOG_LEVEL,
		TlsKeyFile:  DEFAULT_TLS_KEY_FILE,
		TlsCertFile: DEFAULT_TLS_CERT_FILE,
	}

	if config := GetConfig(); !reflect.DeepEqual(*config, want) {
		t.Errorf("GetConfig() = %v, wanted %v", config, want)
	}
}
