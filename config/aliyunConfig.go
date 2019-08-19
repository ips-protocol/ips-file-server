package config

type AliyunConfig struct {
	Endpoint             string `yaml:"endpoint"`
	AccessKey            string `yaml:"accessKey"`
	AccessSecret         string `yaml:"accessSecret"`
	Bucket               string `yaml:"bucket"`
	OssLocation          string `yaml:"ossLocation"`
	Region               string `yaml:"region"`
	MTSPipelineID        string `yaml:"mtsPipelineId"`
	MTSConvertTemplateId string `yaml:"mtsConvertTemplateId"`
}
