package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {

	fmt.Println("здарова мир")

	//создание tcp сервера
	listener, err := net.Listen("tcp4", ":8888") // tcp socket(ipv4)
	if err != nil {
		fmt.Println("err Listener", err) //gg
		return
	}
	defer listener.Close()

	fmt.Println("listener start")

	for {
		conn, err := listener.Accept() //ожидание вход соединения
		if err != nil {
			fmt.Println("err conn", err)
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

	parts := strings.Split(requestLine, " ") //делит строку
	if len(parts) < 2 {
		fmt.Println("invalid request")
		return
	}

	method := parts[0]
	path := parts[1]

	body := fmt.Sprintf("<html><body><h1>Hello! chuvak %s with %s</h1></body></html>", path, method) //создаем строку (тело ответа)

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
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
