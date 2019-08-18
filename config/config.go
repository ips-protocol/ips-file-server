package config

import (
	"github.com/ipweb-group/go-sdk/conf"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	NodeConf conf.Config    `yaml:"node_conf"`
	Redis    RedisConfig    `yaml:"redis"`
	Mongo    MongoConfig    `yaml:"mongo"`
	OSS      OssConfig      `yaml:"OSS"`
	External ExternalConfig `yaml:"external"`
	Clients  []AppClient    `yaml:"clients"`
}

// 配置缓存
var configCache Config

// 加载配置信息
func LoadConfig(configFilePath string) Config {
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Load config file `%s` failed, does this is file exists?", configFilePath)
	}

	err = yaml.Unmarshal(configData, &configCache)
	if err != nil {
		log.Fatalf("Parse config file failed, maybe the format is invalid")
	}

	return configCache
}

// 获取配置信息
func GetConfig() Config {
	return configCache
}
