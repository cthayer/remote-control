package config

import (
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	want := Config{
		Port:    DEFAULT_PORT,
		Host:    DEFAULT_HOST,
		CertDir: DEFAULT_CERT_DIR,
		Ciphers: DEFAULT_CIPHERS,
		PidFile: "",
		EngineOptions: EngineOptions{
			PingTimeout:  DEFAULT_ENGINE_OPTIONS_PING_TIMEOUT,
			PingInterval: DEFAULT_ENGINE_OPTIONS_PING_INTERVAL,
		},
		LogLevel: DEFAULT_LOG_LEVEL,
	}

	if config := GetConfig(); !reflect.DeepEqual(*config, want) {
		t.Errorf("GetConfig() = %v, wanted %v", config, want)
	}
}
