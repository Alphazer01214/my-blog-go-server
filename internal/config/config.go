package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host" json:"host"`
	Port string `yaml:"port" json:"port"`
	Mode string `yaml:"mode" json:"mode"`

	TLSCert         string `yaml:"tls_cert" json:"tls_cert"`
	TLSKey          string `yaml:"tls_key" json:"tls_key"`
	UploadDir       string `yaml:"upload_dir" json:"upload_dir"`
	UploadChunkSize int64  `yaml:"upload_chunk_size" json:"upload_chunk_size"`
}

type Postgres struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Name     string `yaml:"name" json:"name"`

	// 连接池配置
	MaxOpenConns    int `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // 秒
	ConnMaxIdleTime int `yaml:"conn_max_idle_time" json:"conn_max_idle_time"` // 秒
}

type Redis struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	// Timeout seconds
	Timeout  int `yaml:"timeout" json:"timeout"`
	PoolSize int `yaml:"pool_size" json:"pool_size"`
}

type JWT struct {
	AccessTokenSecret  string `yaml:"access_token_secret" json:"access_token_secret"`
	RefreshTokenSecret string `yaml:"refresh_token_secret" json:"refresh_token_secret"`
	// seconds
	AccessTokenExpireTime  int    `yaml:"access_token_expire_time" json:"access_token_expire_time"`
	RefreshTokenExpireTime int    `yaml:"refresh_token_expire_time" json:"refresh_token_expire_time"`
	Issuer                 string `yaml:"issuer" json:"issuer"`
}

type LLM struct {
	Provider  string `yaml:"provider" json:"provider"`
	BaseUrl   string `yaml:"base_url" json:"base_url"`
	ApiKey    string `yaml:"api_key" json:"api_key"`
	ModelName string `yaml:"model_name" json:"model_name"`
}

type Market struct {
	RefreshInterval  int    `yaml:"refresh_interval" json:"refresh_interval"`
	ApiUrl           string `yaml:"api_url" json:"api_url"`
	ExchangeRateUrl  string `yaml:"exchange_rate_url" json:"exchange_rate_url"`
}

// WebSearchConfig 用于配置文件解析，避免循环导入
type WebSearchConfig struct {
	ApiKey string `yaml:"api_key" json:"api_key"`
}

type TushareConfig struct {
	DataDir string `yaml:"data_dir" json:"data_dir"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers" json:"brokers"`
	Enabled bool     `yaml:"enabled" json:"enabled"`
}

type Elasticsearch struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	Index     string   `yaml:"index" json:"index"` // 默认帖子索引名
}

type Tools struct {
	WebSearch WebSearchConfig `yaml:"web_search" json:"web_search"`
	Tushare   TushareConfig   `yaml:"tushare" json:"tushare"`
}

type Config struct {
	Server        *Server        `yaml:"server" json:"server"`
	Postgres      *Postgres      `yaml:"postgres" json:"postgres"`
	LLM           *LLM           `yaml:"llm" json:"llm"`
	Redis         *Redis         `yaml:"redis" json:"redis"`
	JWT           *JWT           `yaml:"jwt" json:"jwt"`
	Market        *Market        `yaml:"market" json:"market"`
	Kafka         *Kafka         `yaml:"kafka" json:"kafka"`
	Elasticsearch *Elasticsearch `yaml:"elasticsearch" json:"elasticsearch"`
	Tools         *Tools         `yaml:"tools" json:"tools"`
}

func (dc *Postgres) GetDSN() string {
	return "host=" + dc.Host + " user=" + dc.User + " password=" + dc.Password + " dbname=" + dc.Name + " port=" + dc.Port + " sslmode=disable"
}

func getEnv(k string, d string) string {
	var val string
	if val = os.Getenv(k); val != "" {
		return val
	}
	return d
}

func LoadConfig() *Config {
	cfgFile, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		panic(err)
	}
	cfg := Config{
		Server:        &Server{},
		Postgres:      &Postgres{},
		LLM:           &LLM{},
		Redis:         &Redis{},
		JWT:           &JWT{},
		Market:        &Market{},
		Kafka:         &Kafka{},
		Elasticsearch: &Elasticsearch{},
	}

	if err := yaml.Unmarshal(cfgFile, &cfg); err != nil {
		panic(err)
	}

	if cfg.Postgres == nil {
		panic("invalid config: postgres section is missing")
	}
	if cfg.Postgres.Host == "" || cfg.Postgres.Port == "" || cfg.Postgres.User == "" || cfg.Postgres.Name == "" {
		panic(fmt.Sprintf("invalid config: postgres fields are incomplete: host=%q port=%q user=%q name=%q", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Name))
	}
	//if err := godotenv.Load(); err != nil {
	//	panic(err)
	//}
	//dbCfg := &Postgres{
	//	Host:     getEnv("DB_HOST", "localhost"),
	//	Port:     getEnv("DB_PORT", "5432"),
	//	User:     getEnv("DB_USER", "postgres"),
	//	Password: getEnv("DB_PASS", ""),
	//	Name:     getEnv("DB_NAME", "tomori_db"),
	//}
	//cfg.PostgreSQL = dbCfg
	return &cfg
}

//func loadServerConfig(path string) *Server {
//	cfg, err := os.ReadFile(path)
//	if err != nil {
//		panic(err)
//	}
//	var ServerCfg Server
//	if err := yaml.Unmarshal(cfg, &ServerCfg); err != nil {
//		panic(err)
//	}
//}
