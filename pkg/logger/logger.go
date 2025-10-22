package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

const maxSize = 5 * 1024 * 1024

var Log *slog.Logger
var rw *rotatingWriter

type rotatingWriter struct {
	file     *os.File
	filepath string
}

func newRotatingWriter(filepath string) (*rotatingWriter, error) {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &rotatingWriter{file: f, filepath: filepath}, nil
}

func (rw *rotatingWriter) Write(p []byte) (n int, err error) {
	info, err := rw.file.Stat()
	if err != nil {
		return 0, err
	}

	if info.Size()+int64(len(p)) > maxSize {
		rw.file.Close()
		backupName := fmt.Sprintf("%s.%s", rw.filepath, time.Now().Format("20060102_150405"))
		os.Rename(rw.filepath, backupName)

		rw.file, err = os.OpenFile(rw.filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
	}

	return rw.file.Write(p)
}

// InitLogger инициализирует глобальный логгер
func InitLogger(logFile, logLevel string) error {
	var err error
	rw, err = newRotatingWriter(logFile)
	if err != nil {
		return err
	}

	var lvl slog.Level
	switch logLevel {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewTextHandler(io.MultiWriter(rw, os.Stdout), &slog.HandlerOptions{
		Level: lvl,
	})

	Log = slog.New(handler)
	return nil
}

// Close закрывает файл лога
func CloseLogger() {
	if rw != nil {
		rw.file.Close()
	}
}
