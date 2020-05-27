package config

type Config struct {
	Port          int           `json:"port"`
	Host          string        `json:"host"`
	CertDir       string        `json:"certDir"`
	Ciphers       string        `json:"ciphers"`
	PidFile       string        `json:"pidFile"`
	EngineOptions EngineOptions `json:"engineOptions"`
	LogLevel      string        `json:"logLevel"`
}

type EngineOptions struct {
	PingTimeout  int `json:"pingTimeout"`
	PingInterval int `json:"pingInterval"`
}

const (
	DEFAULT_PORT                         = 4515
	DEFAULT_HOST                         = ""
	DEFAULT_CERT_DIR                     = "/etc/rc/certs"
	DEFAULT_CIPHERS                      = "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4"
	DEFAULT_ENGINE_OPTIONS_PING_TIMEOUT  = 1000
	DEFAULT_ENGINE_OPTIONS_PING_INTERVAL = 5000
	DEFAULT_LOG_LEVEL                    = "info"
)

var config Config = Config{
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

func GetConfig() *Config {
	return &config
}
