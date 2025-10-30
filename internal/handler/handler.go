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
	"web-server/pkg/jwt"
	"web-server/pkg/logger"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Query   map[string]string
	Body    []byte
	Uploads []*UploadReq
	Headers map[string]string
}

type UploadReq struct {
	Filename    string
	Size        int
	ContentType string
	Content     string
}

// parseRequest читает HTTP-запрос из conn и возвращает Request
func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	// Читаем первую строку запроса: Method Path Version
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Query:   make(map[string]string),
		Headers: make(map[string]string),
	}

	// Читаем заголовки
	contentLength := 0
	contentType := ""
	for {
		hline, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header line: %w", err)
		}
		hline = strings.TrimRight(hline, "\r\n")
		if hline == "" {
			break // конец заголовков
		}

		parts := strings.SplitN(hline, ":", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		req.Headers[name] = value // <-- сохраняем заголовок

		if strings.ToLower(name) == "content-length" {
			if cl, err := strconv.Atoi(value); err == nil {
				contentLength = cl
			}
		}
		if strings.ToLower(name) == "content-type" {
			contentType = value
		}
	}

	// Читаем тело
	if contentLength > 0 {
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.Body = body
	}

	// Если multipart/form-data — разбираем файлы
	if strings.Contains(strings.ToLower(contentType), "multipart/form-data") { //прочитать поподробнее
		if err := parseMultipart(req, contentType); err != nil {
			return nil, err
		}
	}

	// Парсим query-параметры
	if idx := strings.Index(req.Path, "?"); idx != -1 {
		rawQuery := req.Path[idx+1:]
		req.Path = req.Path[:idx]
		for _, pair := range strings.Split(rawQuery, "&") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				req.Query[kv[0]] = kv[1]
			}
		}
	} //проверить спец символы

	fmt.Println("\nHeaders:")
	for k, v := range req.Headers {
		fmt.Printf("%-20s : %s\n", k, v)
	}

	if len(req.Body) > 0 {
		fmt.Printf("тело запроса (%d байт):\n%s\n", len(req.Body), string(req.Body))
	} else {
		fmt.Println("тело запроса: <пустое>")
	}
	fmt.Println(strings.Repeat("-", 80))

	return req, nil
}

func parseMultipart(req *Request, contentType string) error {
	boundaryIdx := strings.Index(contentType, "boundary=")
	if boundaryIdx == -1 {
		return fmt.Errorf("missing boundary in multipart/form-data")
	}

	boundary := "--" + contentType[boundaryIdx+9:]
	parts := strings.Split(string(req.Body), boundary)

	for _, part := range parts {
		part = strings.Trim(part, "\r\n")
		if part == "" || part == "--" {
			continue
		}

		lines := strings.Split(part, "\n")
		var filename, partContentType string
		ctStart := 0

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				ctStart = i + 1
				break
			}
			if strings.HasPrefix(line, "Content-Disposition:") {
				if idx := strings.Index(line, "filename="); idx != -1 {
					filename = strings.Trim(line[idx+9:], "\"")
					filename = strings.TrimSpace(filename)
				}
			}
			if strings.HasPrefix(strings.ToLower(line), "content-type:") {
				partContentType = strings.TrimSpace(line[len("Content-Type:"):])
			}
		}

		if partContentType == "" {
			partContentType = "application/octet-stream"
		}

		fcontent := strings.Join(lines[ctStart:], "\n")
		uploadF := &UploadReq{
			Filename:    filename,
			Size:        len(fcontent),
			ContentType: partContentType,
			Content:     fcontent,
		}

		req.Uploads = append(req.Uploads, uploadF)
	}

	return nil
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

		if _, err := jwt.ParseToken(cfg, req.Headers); err != nil {
			fmt.Print("недействительный токен", "error", err)
			sendStatus(conn, 401)
			return
		}

		var u model.User
		if err := json.Unmarshal(req.Body, &u); err != nil {
			logger.Log.Warn("некорректное тело запроса", "error", err)
			sendStatus(conn, 400)
			return
		}
		hashed, err := store.HashPassword(u.Password)
		if err != nil {
			logger.Log.Error("ошибка хэширования пароля", "error", err)
			sendStatus(conn, 500)
			return
		}
		u.Password = hashed
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

	case req.Method == "POST" && req.Path == base+"/users/upload":
		if len(req.Uploads) == 0 {
			logger.Log.Warn("нет загруженных файлов")
			sendStatus(conn, 400)
			return
		}

		sendJSON(conn, 200, req.Uploads)

	case req.Method == "POST" && req.Path == base+"/users/login":

		var creds struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}

		if err := json.Unmarshal(req.Body, &creds); err != nil {
			logger.Log.Warn("некорректное тело запроса", "error", err)
			sendStatus(conn, 400)
			return
		}

		user, ok := store.GetUserByLogin(creds.Login)
		if !ok {
			logger.Log.Warn("пользователь не найден", "login", creds.Login)
			sendStatus(conn, 401)
			return
		}

		if err := store.VerifyPassword(user.Password, creds.Password); err != nil {
			logger.Log.Warn("неверный пароль", "login", creds.Login)
			sendStatus(conn, 401)
			return
		}

		token, err := jwt.GenerateToken(user.ID, cfg)
		if err != nil {
			logger.Log.Error("ошибка генерации токена", "error", err)
			sendStatus(conn, 500)
			return
		}

		resp := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
			"access_token": token,
		}

		sendJSON(conn, 200, resp)
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
