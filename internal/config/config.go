package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string        `yaml:"env" env-default:"local"`
	HTTP    HTTPConfig    `yaml:"http"`
	Clients ClientsConfig `yaml:"clients"`
}

type HTTPConfig struct {
	Port    int           `yaml:"port" env:"HTTP_PORT"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type ClientsConfig struct {
	SSO    SSOConfig    `yaml:"sso"`
	Seller SellerConfig `yaml:"seller"`
}

type SSOConfig struct {
	Address string        `yaml:"address" env:"SSO_ADDRESS" env-required:"true"`
	AppId   int32         `yaml:"app_id" env:"SSO_APP_ID" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type SellerConfig struct {
	Address string        `yaml:"address" env:"SELLER_ADDRESS" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
