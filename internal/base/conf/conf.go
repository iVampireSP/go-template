package conf

// Config 配置文件不能有下划线或横线，否则不能解析
type Config struct {
	App *App `yaml:"app"`

	Http *Http `yaml:"http"`

	Grpc *Grpc `yaml:"grpc"`

	Debug *Debug `yaml:"debug"`

	Database *Database `yaml:"database"`

	Redis *Redis `yaml:"redis"`

	JWKS *JWKS `yaml:"jwks"`

	Metrics *Metrics `yaml:"metrics"`

	S3 *S3 `yaml:"s3"`

	Kafka *Kafka `yaml:"kafka"`

	Milvus *Milvus `yaml:"milvus"`

	ThirdParty *ThirdParty `yaml:"third_party" mapstructure:"third_party"`
}

type App struct {
	Name             string   `yaml:"name"`
	AllowedAudiences []string `yaml:"allowed_audiences" mapstructure:"allowed_audiences"`
}

type ThirdParty struct {
	OpenAIApiKey string `yaml:"openai_api_key" mapstructure:"openai_api_key"`
}

type Http struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
	Url  string `yaml:"url" mapstructure:"url"`
	Cors struct {
		Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
		AllowedOrigins   []string `yaml:"allow_origins" mapstructure:"allow_origins"`
		AllowMethods     []string `yaml:"allow_methods" mapstructure:"allow_methods"`
		AllowHeaders     []string `yaml:"allow_headers" mapstructure:"allow_headers"`
		AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
		ExposeHeaders    []string `yaml:"expose_headers" mapstructure:"expose_headers"`
		MaxAge           int      `yaml:"max_age" mapstructure:"max_age"`
	} `yaml:"cors" mapstructure:"cors"`
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
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
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
	BootstrapServers []string `yaml:"bootstrap_servers" mapstructure:"bootstrap_servers"`
	GroupId          string   `yaml:"group_id" mapstructure:"group_id"`
	Username         string   `yaml:"username" mapstructure:"username"`
	Password         string   `yaml:"password" mapstructure:"password"`
	MainTopic        string   `yaml:"main_topic" mapstructure:"main_topic"`
	WorkerTopic      string   `yaml:"worker_topic" mapstructure:"worker_topic"`
}

type Milvus struct {
	Host               string `yaml:"host" mapstructure:"host"`
	Port               int    `yaml:"port" mapstructure:"port"`
	DBName             string `yaml:"db_name" mapstructure:"db_name"`
	DocumentCollection string `yaml:"document_collection" mapstructure:"document_collection"`
	User               string `yaml:"user" mapstructure:"user"`
	Password           string `yaml:"password" mapstructure:"password"`
}

type Grpc struct {
	Address        string `yaml:"address" mapstructure:"address"`
	AddressGateway string `yaml:"address_gateway" mapstructure:"address_gateway"`
}
