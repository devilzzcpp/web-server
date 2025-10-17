package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// loadCfg загружает конфигурацию из файла
func LoadCfg() (*Config, error) {

	// открываем файл кфг
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//декодирование
	cfg := &Config{}
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
