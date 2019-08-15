package config

type MongoConfig struct {
	ConnectionUri string `yaml:"connection_uri"`
	Db            string `yaml:"db"`
}
