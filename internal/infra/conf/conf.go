package conf

// Config 配置文件不能有下划线或横线，否则不能解析
type Config struct {
	Listen *Listen `yaml:"listen"`

	Debug *Debug `yaml:"debug"`

	Database *Database `yaml:"database"`

	Redis *Redis `yaml:"redis"`

	S3 *S3 `yaml:"s3"`

	Kafka *Kafka `yaml:"kafka"`
}

type Listen struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
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

//type Milvus struct {
//	Host               string `yaml:"host" mapstructure:"host"`
//	Port               int    `yaml:"port" mapstructure:"port"`
//	DBName             string `yaml:"db_name" mapstructure:"db_name"`
//	DocumentCollection string `yaml:"document_collection" mapstructure:"document_collection"`
//	User               string `yaml:"user" mapstructure:"user"`
//	Password           string `yaml:"password" mapstructure:"password"`
//}
