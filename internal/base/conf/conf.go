package conf

// Config 配置文件不能有下划线或横线，否则不能解析
type Config struct {
	Http *Http `yaml:"http"`

	Grpc *Grpc `yaml:"grpc"`

	Debug *Debug `yaml:"debug"`

	Database *Database `yaml:"database"`

	Redis *Redis `yaml:"redis"`

	JWKS *JWKS `yaml:"jwks"`

	Metrics *Metrics `yaml:"metrics"`

	S3 *S3 `yaml:"s3"`

	Kafka *Kafka `yaml:"kafka"`

	ThirdParty *ThirdParty `yaml:"third_party" mapstructure:"third_party"`
}

type ThirdParty struct {
	OpenAIApiKey string `yaml:"openai_api_key" mapstructure:"openai_api_key"`
}

type Http struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Url  string `yaml:"url"`
}

type Debug struct {
	Enabled bool `yaml:"enabled"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWKS struct {
	Url string `yaml:"url"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
}

type S3 struct {
	Endpoint         string `yaml:"endpoint" mapstructure:"endpoint"`
	ExternalEndpoint string `yaml:"external_endpoint" mapstructure:"external_endpoint"`
	AccessKey        string `yaml:"access_key" mapstructure:"access_key"`
	SecretKey        string `yaml:"secret_key" mapstructure:"secret_key"`
	Bucket           string `yaml:"bucket" mapstructure:"bucket"`
	UseSSL           bool   `yaml:"use_ssl" mapstructure:"use_ssl"`
	Region           string `yaml:"region" mapstructure:"region"`
}

type Kafka struct {
	BootstrapServers KafkaBootstrapServers `yaml:"bootstrap_servers" mapstructure:"bootstrap_servers"`
	Topic            string                `yaml:"topic" mapstructure:"topic"`
	GroupId          string                `yaml:"group_id" mapstructure:"group_id"`
	Username         string                `yaml:"username" mapstructure:"username"`
	Password         string                `yaml:"password" mapstructure:"password"`
}

type KafkaBootstrapServers []string

type Grpc struct {
	Address string `yaml:"address" mapstructure:"address"`
}
