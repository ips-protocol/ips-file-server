package config

type ServerConfig struct {
	HttpHost     string `yaml:"http_host"`
	EnableHttps  bool   `yaml:"enable_https"`
	HttpsHost    string `yaml:"https_host"`
	HttpsDomains string `yaml:"https_domains"`
	HttpsEmail   string `yaml:"https_email"`
}
