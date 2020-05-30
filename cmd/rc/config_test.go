package main

import (
	"reflect"
	"testing"

	config "github.com/cthayer/remote_control/pkg/client_config"
)

func TestCliConf(t *testing.T) {
	want := cliConfig{
		ConfigFile:    DEFAULT_CLI_CONF_CONFIG_FILE,
		Port:          config.DEFAULT_PORT,
		KeyDir:        config.DEFAULT_KEY_DIR,
		KeyName:       config.DEFAULT_KEY_NAME,
		LogLevel:      config.DEFAULT_LOG_LEVEL,
		BatchSize:     DEFAULT_CLI_CONF_BATCH_SIZE,
		Delay:         DEFAULT_CLI_CONF_DELAY,
		Verbose:       DEFAULT_CLI_CONF_VERBOSE,
		Retry:         DEFAULT_CLI_CONF_RETRY,
		TlsCaFile:     config.DEFAULT_TLS_CA_FILE,
		TlsSkipVerify: config.DEFAULT_TLS_SKIP_VERIFY,
		TlsDisable:    config.DEFAULT_TLS_DISABLE,
	}

	if !reflect.DeepEqual(cliConf, want) {
		t.Errorf("GetConfig() = %v, wanted %v", cliConf, want)
	}
}
