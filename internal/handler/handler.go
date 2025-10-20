package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"web-server/internal/model"
	"web-server/internal/storage"
)

func HandleConnection(conn net.Conn, storage *storage.Storage) {

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n') //читает строку запроса
	if err != nil {
		fmt.Println("err read", err)
		return
	}

	requestLine = strings.TrimSpace(requestLine)
	fmt.Println("Received:", requestLine)

	// читаем заголовки до пустой строки
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("err read header:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
	}

	parts := strings.Fields(requestLine) //делит строку
	if len(parts) != 3 {
		fmt.Println("invalid request")
		return
	}

	method := parts[0]
	path := parts[1]
	version := parts[2]

	fmt.Printf("Method=%s, Path=%s, Version=%s\n", method, path, version)

	if method == "GET" && strings.HasPrefix(path, "/api/v1/users") {
		role := ""
		if idx := strings.Index(path, "?"); idx != -1 {
			role = path[idx+1:]
		}

		var users []model.User
		if role == "" {
			users = storage.GetUsers()
		} else {
			users = storage.GetUsersByRole(role)
		}

		body, _ := json.Marshal(users)

		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s", len(body), string(body))
		conn.Write([]byte(response))
		return
	}

	if method == "POST" && path == "/api/v1/users" {
		body, _ := io.ReadAll(reader)
		var user model.User
		json.Unmarshal(body, &user)

		newUser := storage.CreateUser(user)
		respByte, _ := json.Marshal(newUser)

		response := fmt.Sprintf("HTTP/1.1 201 Created\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s", len(respByte), string(respByte))
		conn.Write([]byte(response))
		return
	}

	if method == "PUT" && strings.HasPrefix(path, "/api/v1/users/") {
		strId := strings.TrimPrefix(path, "/api/v1/users/")

		id, err := strconv.Atoi(strId)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		body, _ := io.ReadAll(reader)
		var user model.User
		json.Unmarshal(body, &user)

		updUser, ok := storage.UpdateUser(id, user)
		if !ok {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		respByte, _ := json.Marshal(updUser)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s", len(respByte), string(respByte))
		conn.Write([]byte(response))
		return
	}

	if method == "DELETE" && strings.HasPrefix(path, "/api/v1/users/") {
		strID := strings.TrimPrefix(path, "/api/v1/users/")
		id, err := strconv.Atoi(strID)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		ok := storage.DeleteUser(id)
		if !ok {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		conn.Write([]byte("HTTP/1.1 204 No Content\r\nContent-Length: 0\r\n\r\n"))
		return
	}

	if method == "GET" && strings.HasPrefix(path, "/api/v1/users/") {
		strID := strings.TrimPrefix(path, "/api/v1/users/")
		id, err := strconv.Atoi(strID)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		user, ok := storage.GetUser(id)
		if !ok {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Length: 0\r\n\r\n"))
			return
		}

		respByte, _ := json.Marshal(user)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s", len(respByte), string(respByte))
		conn.Write([]byte(response))
		return
	}

}
