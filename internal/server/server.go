package server

import (
	"fmt"
	"net"
	"web-server/internal/config"
	"web-server/internal/handler"
	"web-server/internal/storage"
	"web-server/pkg/logger"
)

func Start(cfg *config.Config, storage *storage.Storage) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		logger.Log.Error("ошибка запуска listener", "error", err)
		return
	}
	defer listener.Close()

	logger.Log.Info("listener запущен на", "address", addr)

	for {
		conn, err := listener.Accept() //ожидание вход соединения
		if err != nil {
			logger.Log.Error("ошибка принятия соединения", "error", err)
			return
		}
		logger.Log.Info("новое подключение", "address", conn.RemoteAddr())
		go handler.HandleConnection(conn, storage, cfg)
	}

}
