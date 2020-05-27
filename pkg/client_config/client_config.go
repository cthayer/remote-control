package client_config

type Config struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	KeyDir   string `json:"keyDir"`
	KeyName  string `json:"keyName"`
	LogLevel string `json:"logLevel"`
}

const (
	DEFAULT_PORT      = 4515
	DEFAULT_HOST      = "localhost"
	DEFAULT_KEY_DIR   = ""
	DEFAULT_KEY_NAME  = ""
	DEFAULT_LOG_LEVEL = "info"
)

var config Config = Config{
	Port:     DEFAULT_PORT,
	Host:     DEFAULT_HOST,
	KeyDir:   DEFAULT_KEY_DIR,
	KeyName:  DEFAULT_KEY_NAME,
	LogLevel: DEFAULT_LOG_LEVEL,
}

func GetConfig() *Config {
	return &config
}
