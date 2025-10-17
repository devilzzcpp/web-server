package main

import (
	"fmt"
	"web-server/internal/config"
	"web-server/internal/server"
)

func main() {
	cfg, err := config.LoadCfg()
	if err != nil {
		fmt.Printf("ошибка загрузки конфига: %v\n", err)
		return
	}

	fmt.Printf("запуск сервера %s:%d\n", cfg.Host, cfg.Port)

	server.Start(cfg)
}
