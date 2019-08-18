package config

type OssConfig struct {
	Endpoint     string `yaml:"endpoint"`
	AccessKey    string `yaml:"accessKey"`
	AccessSecret string `yaml:"accessSecret"`
	Bucket       string `yaml:"bucket"`
}
