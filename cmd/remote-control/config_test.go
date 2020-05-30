package main

import (
	"reflect"
	"testing"

	"github.com/cthayer/remote_control/internal/config"
)

func TestCliConf(t *testing.T) {
	want := cliConfig{
		ConfigFile:  DEFAULT_CLI_CONF_CONFIG_FILE,
		Port:        config.DEFAULT_PORT,
		CertDir:     config.DEFAULT_CERT_DIR,
		Ciphers:     config.DEFAULT_CIPHERS,
		LogLevel:    config.DEFAULT_LOG_LEVEL,
		Host:        config.DEFAULT_HOST,
		PidFile:     DEFAULT_CLI_CONF_PID_FILE,
		TlsKeyFile:  config.DEFAULT_TLS_KEY_FILE,
		TlsCertFile: config.DEFAULT_TLS_CERT_FILE,
	}

	if !reflect.DeepEqual(cliConf, want) {
		t.Errorf("GetConfig() = %v, wanted %v", cliConf, want)
	}
}
