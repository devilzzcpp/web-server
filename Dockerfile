# Используем минимальный Go 1.23 образ
FROM golang:1.23-alpine

# Рабочая директория внутри контейнера
WORKDIR /app

# Копируем только файлы модулей и скачиваем зависимости для кеширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение
RUN go build -o web-server ./cmd/server

# Открываем порт, который использует сервер
EXPOSE 8888

# Создаём папку для логов
RUN mkdir -p /app/logs

# Команда запуска
CMD ["./web-server"]
