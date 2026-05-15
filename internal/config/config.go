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

	TLSCert string `yaml:"tls_cert" json:"tls_cert"`
	TLSKey  string `yaml:"tls_key" json:"tls_key"`
}

type Postgres struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Name     string `yaml:"name" json:"name"`
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
	RefreshInterval int    `yaml:"refresh_interval" json:"refresh_interval"`
	ApiUrl          string `yaml:"api_url" json:"api_url"`
}

type Config struct {
	Server   *Server   `yaml:"server" json:"server"`
	Postgres *Postgres `yaml:"postgres" json:"postgres"`
	LLM      *LLM      `yaml:"llm" json:"llm"`
	Redis    *Redis    `yaml:"redis" json:"redis"`
	JWT      *JWT      `yaml:"jwt" json:"jwt"`
	Market   *Market   `yaml:"market" json:"market"`
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
		Server:   &Server{},
		Postgres: &Postgres{},
		LLM:      &LLM{},
		Redis:    &Redis{},
		JWT:      &JWT{},
		Market:   &Market{},
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
