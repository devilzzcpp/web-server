# Используем минимальный Go 1.23 образ
FROM golang:1.23-alpine AS builder

# Устанавливаем нужные инструменты
RUN apk add --no-cache git ca-certificates

# Рабочая директория
WORKDIR /app

# Копируем модули и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o web-server ./cmd/server


FROM alpine:3.18

# Устанавливаем сертификаты для HTTPS (если нужно)
RUN apk add --no-cache ca-certificates

# Рабочая директория
WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/web-server ./

# Создаём папку для логов
RUN mkdir -p /app/logs

# Порт из переменной окружения (по умолчанию 8886)
ENV PORT=8886
EXPOSE $PORT

# Дефолтная команда запуска
CMD ["sh", "-c", "./web-server --port=$PORT"]
