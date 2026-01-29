package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gladinov/e"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env          string       `env:"ENV" env-required:"true"`
	RootPath     string       `env:"ROOT_PATH" env-required:"true"`
	ConfigPath   string       `env:"CONFIG_PATH" env-required:"true"`
	PostgresHost PostgresHost `yaml:"postgresHost"`
	Clients      Clients      `yaml:"clients"`
}

type Clients struct {
	TestAppClient TestAppClient `yaml:"testAppClient"`
}

type TestAppClient struct {
	Host string `yaml:"testAppHost"`
	Port string `env:"TEST_APP_PORT" env-required:"true"`
}

func (c *TestAppClient) GetTestAppClientAddress() string {
	return getAddress(c.Host, c.Port)
}

type PostgresHost struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Dbname   string `env:"POSTGRES_DB" env-required:"true"`
	Port     string `env:"POSTGRES_SERVICE_PORT" env-required:"true"`
	PgUser   string `env:"PGUSER" env-required:"true"`
	SslMode  string `yaml:"sslmode"`
}

func (p *PostgresHost) GetStringHost() (string, error) {
	if p.Host == "" {
		return "", errors.New("empty host in config")
	}
	if p.User == "" {
		return "", errors.New("empty user in config")
	}
	if p.Password == "" {
		return "", errors.New("empty password in config")
	}
	if p.Dbname == "" {
		return "", errors.New("empty dbname in config")
	}
	if p.Port == "" {
		return "", errors.New("empty port in config")
	}
	host := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=%s",
		p.Host,
		p.User,
		p.Password,
		p.Dbname,
		p.Port,
		p.SslMode)
	return host, nil
}

func InitConfig() (Config, error) {
	const op = "config.InitConfig"
	envs, err := InjectEnvs()
	if err != nil {
		return Config{}, e.Wrap("InjectEnvs failed", err)
	}

	path := filepath.Join(envs.RootPath, envs.ConfigPath)
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return Config{}, e.Wrap("cannot read config", err)
	}

	return cfg, nil
}

func MustInitConfig() Config {
	const op = "config.MustInitConfig"
	cfg, err := InitConfig()
	if err != nil {
		log.Fatalf("%s", err)
	}
	return cfg
}

type Envs struct {
	RootPath   string
	ConfigPath string
}

func InjectEnvs() (Envs, error) {
	rootPath := os.Getenv("ROOT_PATH")
	if rootPath == "" {
		return Envs{}, errors.New("ROOT_PATH environment variable is required")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return Envs{}, errors.New("CONFIG_PATH environment variable is required")
	}

	envs := Envs{
		RootPath:   rootPath,
		ConfigPath: configPath,
	}

	return envs, nil
}

func getAddress(host string, port string) string {
	return host + ":" + port
}
