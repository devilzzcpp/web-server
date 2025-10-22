package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host        string
	Port        int
	LogLevel    string
	LogFile     string
	ApiBasePath string
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

	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "INFO"
	}
	cfg.LogLevel = loglevel

	logfile := os.Getenv("LOG_FILE")
	if logfile == "" {
		logfile = "server.log"
	}
	cfg.LogFile = logfile

	ApiBasePath := os.Getenv("API_BASE_PATH")
	if ApiBasePath == "" {
		ApiBasePath = "/api/v1"
	}
	cfg.ApiBasePath = ApiBasePath

	return cfg, nil
}
