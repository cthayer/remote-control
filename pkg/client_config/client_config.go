package client_config

type Config struct {
	Port          int    `json:"port"`
	Host          string `json:"host"`
	KeyDir        string `json:"keyDir"`
	KeyName       string `json:"keyName"`
	LogLevel      string `json:"logLevel"`
	TlsSkipVerify bool   `json:"tlsSkipVerify"`
	TlsCaFile     string `json:"tlsCaFile"`
	TlsDisable    bool   `json:"tlsDisable"`
}

const (
	DEFAULT_PORT            = 4515
	DEFAULT_HOST            = "localhost"
	DEFAULT_KEY_DIR         = ""
	DEFAULT_KEY_NAME        = ""
	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_TLS_CA_FILE     = ""
	DEFAULT_TLS_SKIP_VERIFY = false
	DEFAULT_TLS_DISABLE     = false
)

var config Config = Config{
	Port:          DEFAULT_PORT,
	Host:          DEFAULT_HOST,
	KeyDir:        DEFAULT_KEY_DIR,
	KeyName:       DEFAULT_KEY_NAME,
	LogLevel:      DEFAULT_LOG_LEVEL,
	TlsSkipVerify: DEFAULT_TLS_SKIP_VERIFY,
	TlsCaFile:     DEFAULT_TLS_CA_FILE,
	TlsDisable:    DEFAULT_TLS_DISABLE,
}

func GetConfig() *Config {
	return &config
}
