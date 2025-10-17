package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host string
	Port int
}

// loadCfg загружает конфигурацию из файла
func LoadCfg() (*Config, error) {

	_ = godotenv.Load() //подгружает env

	cfg := &Config{}

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	cfg.Host = host

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8888"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	cfg.Port = port

	return cfg, nil
}
