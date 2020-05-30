package config

type Config struct {
	Port        int    `json:"port"`
	Host        string `json:"host"`
	CertDir     string `json:"certDir"`
	Ciphers     string `json:"ciphers"`
	LogLevel    string `json:"logLevel"`
	TlsCertFile string `json:"tlsCertFile"`
	TlsKeyFile  string `json:"tlsKeyFile"`
}

type EngineOptions struct {
	PingTimeout  int `json:"pingTimeout"`
	PingInterval int `json:"pingInterval"`
}

const (
	DEFAULT_PORT                         = 4515
	DEFAULT_HOST                         = ""
	DEFAULT_CERT_DIR                     = "/etc/rc/certs"
	DEFAULT_CIPHERS                      = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	DEFAULT_ENGINE_OPTIONS_PING_TIMEOUT  = 1000
	DEFAULT_ENGINE_OPTIONS_PING_INTERVAL = 5000
	DEFAULT_LOG_LEVEL                    = "info"
	DEFAULT_TLS_KEY_FILE                 = ""
	DEFAULT_TLS_CERT_FILE                = ""
)

var config Config = Config{
	Port:        DEFAULT_PORT,
	Host:        DEFAULT_HOST,
	CertDir:     DEFAULT_CERT_DIR,
	Ciphers:     DEFAULT_CIPHERS,
	TlsCertFile: DEFAULT_TLS_CERT_FILE,
	TlsKeyFile:  DEFAULT_TLS_KEY_FILE,
	LogLevel:    DEFAULT_LOG_LEVEL,
}

func GetConfig() *Config {
	return &config
}
