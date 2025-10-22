package main

import (
	"fmt"
	"web-server/internal/config"
	"web-server/internal/server"
	"web-server/internal/storage"
	"web-server/pkg/logger"
)

func main() {
	cfg, err := config.LoadCfg()
	if err != nil {
		fmt.Printf("ошибка загрузки конфига: %v\n", err)
		return
	}

	if err := logger.InitLogger(cfg.LogFile, cfg.LogLevel); err != nil {
		fmt.Printf("ошибка инициализации логгера: %v\n", err)
		return
	}
	defer logger.CloseLogger()

	logger.Log.Info("Конфигурация загружена: ",
		"host", cfg.Host,
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
		"log_file", cfg.LogFile)

	storage := storage.NewStorage()
	storage.SeedUsers()
	logger.Log.Info("Хранилище пользователей инициализировано")

	logger.Log.Info("Запуск сервера", "host", cfg.Host, "port", cfg.Port)
	server.Start(cfg, storage)

}
