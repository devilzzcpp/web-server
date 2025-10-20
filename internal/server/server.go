package server

import (
	"fmt"
	"net"
	"web-server/internal/config"
	"web-server/internal/handler"
	"web-server/internal/storage"
)

func Start(cfg *config.Config, storage *storage.Storage) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		fmt.Println("ошибка запуска listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("listener запущен на", addr)

	for {
		conn, err := listener.Accept() //ожидание вход соединения
		if err != nil {
			fmt.Println("err conn", err)
			return
		}
		go handler.HandleConnection(conn, storage)
	}

}
