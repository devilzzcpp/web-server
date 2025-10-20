package main

import (
	"fmt"
	"web-server/internal/config"
	"web-server/internal/server"
	"web-server/internal/storage"
)

func main() {
	cfg, err := config.LoadCfg()
	if err != nil {
		fmt.Printf("ошибка загрузки конфига: %v\n", err)
		return
	}

	fmt.Printf("запуск сервера %s:%d\n", cfg.Host, cfg.Port)

	storage := storage.NewStorage()
	storage.SeedUsers()

	server.Start(cfg, storage)

}
