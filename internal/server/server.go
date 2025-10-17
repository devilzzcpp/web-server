package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"web-server/internal/config"
)

func Start(cfg *config.Config) {
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
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n') //читает строку запроса
	if err != nil {
		fmt.Println("err read", err)
		return
	}

	requestLine = strings.TrimSpace(requestLine) //убирает всякую хуйню
	fmt.Println("Received:", requestLine)

	//читает до пустой строки
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("err read header", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
	}

	parts := strings.Fields(requestLine) //делит строку
	if len(parts) < 2 {
		fmt.Println("invalid request")
		return
	}

	method := parts[0]
	path := parts[1]
	version := parts[2]

	body := fmt.Sprintf("<html><body><h1>Hello! Метод: %s  Путь: %s  Версия: %s</h1></body></html>", method, path, version) //создаем строку (тело ответа)

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"\r\n" +
		body

	_, err = conn.Write([]byte(response)) //отправляем ответ
	if err != nil {
		fmt.Println("err write", err)
		return
	}

	// defer conn.Close()
	// buffer := make([]byte, 1024)
	// n, _ := conn.Read(buffer)

	// fmt.Println(string(buffer[:n]))

}
