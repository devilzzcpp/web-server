package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"web-server/internal/config"
	"web-server/internal/model"
	"web-server/internal/storage"
	"web-server/pkg/logger"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Query   map[string]string
	Body    []byte
}

// parseRequest читает HTTP-запрос из conn и возвращает Request
func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	//fmt.Print("строка запроса", string(reader))
	// fmt.Println("zhopa")
	// b, _ := reader.Peek(512) // возвращает []byte, не потребляя их
	// fmt.Println(string(b))
	// читаем первую строку запроса
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	version := parts[2]
	if version != "HTTP/1.1" && version != "HTTP/1.0" {
		return nil, fmt.Errorf(" неподдерживаемая версия HTTP: %s", version)
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: version,
		Query:   make(map[string]string),
	}

	// читаем заголовки и ищем Content-Length
	contentLength := 0
	fmt.Println("Headers:")
	for {
		hline, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}
		// убираем только CRLF/ LF, чтобы сохранить возможные двоеточия в значениях
		hline = strings.TrimRight(hline, "\r\n")
		if hline == "" {
			break // конец заголовков
		}

		parts := strings.SplitN(hline, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("Malformed-Header: %s\n", hline)
			continue
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// красивый вывод в терминал
		fmt.Printf("%-20s: %s\n", name, value)

		if strings.ToLower(name) == "content-length" {
			if cl, err := strconv.Atoi(value); err == nil {
				contentLength = cl
			}
		}
	}

	// читаем тело ровно contentLength байт
	if contentLength > 0 {
		body := make([]byte, contentLength)
		n, err := io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.Body = body
		fmt.Printf("тело запроса (%d байт):\n%s\n", n, string(body))
	} else {
		fmt.Println("тело запроса: <пустое>")
	}

	// парсим query-параметры
	// /my-api/v1/test?kqeqwkek&askdas=12313&fkdsfk=12
	if idx := strings.Index(req.Path, "?"); idx != -1 {
		rawQuery := req.Path[idx+1:]
		req.Path = req.Path[:idx]
		for _, pair := range strings.Split(rawQuery, "&") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				req.Query[kv[0]] = kv[1]
			}
		}
	}

	return req, nil
}

// HandleConnection обрабатывает одно TCP соединение
func HandleConnection(conn net.Conn, store *storage.Storage, cfg *config.Config) {
	defer conn.Close()

	logger.Log.Info("новое подключение", "address", conn.RemoteAddr())
	req, err := parseRequest(conn)
	if err != nil {
		logger.Log.Error("ошибка парсинга запроса", "error", err)
		sendStatus(conn, 400)
		return
	}

	logger.Log.Info("получен запрос", "method", req.Method, "path", req.Path, "query", req.Query)

	base := cfg.ApiBasePath

	switch {

	case req.Method == "GET" && req.Path == base+"/users":
		role := req.Query["role"]
		//fmt.Print("role debug", role)
		var users []model.User
		if role == "" {
			users = store.GetUsers()
			logger.Log.Info("получены все пользователи")
		} else {
			users = store.GetUsersByRole(role)
			logger.Log.Info("получены пользователи по роли", "role", role)
		}
		sendJSON(conn, 200, users)
		return

	case req.Method == "GET" && strings.HasPrefix(req.Path, base+"/users/"):
		idStr := strings.TrimPrefix(req.Path, base+"/users/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Warn("некорректный ID пользователя", "error", err)
			sendStatus(conn, 400)
			return
		}
		user, ok := store.GetUser(id)
		if !ok {
			logger.Log.Warn("пользователь не найден", "id", id)
			sendStatus(conn, 404)
			return
		}
		logger.Log.Info("пользователь найден", "id", id)
		sendJSON(conn, 200, user)
		return

	case req.Method == "POST" && req.Path == base+"/users":
		var u model.User
		if err := json.Unmarshal(req.Body, &u); err != nil {
			logger.Log.Warn("некорректное тело запроса", "error", err)
			sendStatus(conn, 400)
			return
		}
		u.ID = 0
		newUser := store.CreateUser(u)
		sendJSON(conn, 201, newUser)
		return

	case req.Method == "PUT" && strings.HasPrefix(req.Path, base+"/users/"):
		idStr := strings.TrimPrefix(req.Path, base+"/users/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Warn("некорректный ID пользователя", "error", err)
			sendStatus(conn, 400)
			return
		}
		var u model.User
		if err := json.Unmarshal(req.Body, &u); err != nil {
			logger.Log.Warn("некорректное тело запроса", "error", err)
			sendStatus(conn, 400)
			return
		}
		updatedUser, ok := store.UpdateUser(id, u)
		if !ok {
			logger.Log.Warn("пользователь не найден", "id", id)
			sendStatus(conn, 404)
			return
		}
		logger.Log.Info("пользователь обновлен", "id", id)
		sendJSON(conn, 200, updatedUser)
		return

	case req.Method == "DELETE" && strings.HasPrefix(req.Path, base+"/users/"):
		idStr := strings.TrimPrefix(req.Path, base+"/users/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.Log.Warn("некорректный ID пользователя", "error", err)
			sendStatus(conn, 400)
			return
		}
		if !store.DeleteUser(id) {
			logger.Log.Warn("пользователь не найден", "id", id)
			sendStatus(conn, 404)
			return
		}
		logger.Log.Info("пользователь удален", "id", id)
		sendStatus(conn, 204)
		return

	default:
		logger.Log.Warn("неизвестный метод или путь", "method", req.Method, "path", req.Path)
		sendStatus(conn, 405)
	}

	logger.Log.Info("обработка запроса завершена", "address", conn.RemoteAddr())
}

// sendJSON отправляет JSON с указанным статусом
func sendJSON(conn net.Conn, status int, data interface{}) {
	body, _ := json.Marshal(data)
	resp := fmt.Sprintf(
		"HTTP/1.1 %d OK\r\nContent-Type: application/json; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s",
		status, len(body), string(body),
	)
	conn.Write([]byte(resp))

	logger.Log.Debug("отправлен JSON-ответ",
		"status", status,
		"length", len(body),
		"address", conn.RemoteAddr(),
	)
}

// sendStatus отправляет пустой ответ с кодом состояния
func sendStatus(conn net.Conn, status int) {
	statusText := map[int]string{
		200: "OK",
		201: "Created",
		204: "No Content",
		400: "Bad Request",
		404: "Not Found",
		405: "Method Not Allowed",
	}[status]
	if statusText == "" {
		statusText = "Unknown"
	}
	resp := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, statusText)
	_, _ = conn.Write([]byte(resp))
}
