package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

type migrateLogger struct{}

func (m migrateLogger) Printf(format string, v ...any) {
	slog.Info(fmt.Sprintf(format, v...))
}

func (m migrateLogger) Verbose() bool {
	return true
}

func appendDSNParam(dsn, key, value string) string {
	sep := "?"
	if strings.ContainsRune(dsn, '?') {
		sep = "&"
	}
	return dsn + sep + key + "=" + value
}

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file")
	}

	//SetupLogger("migrator_ostrovok.log", slog.LevelDebug)

	envStoragePath := os.Getenv("DATABASE_URL")

	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", envStoragePath, "PostgreSQL DSN")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migration")
	flag.Parse()

	if storagePath == "" {
		slog.Error("storage-path is required (PostgreSQL DSN)")
		os.Exit(1)
	}
	if migrationsPath == "" {
		slog.Error("migrations-path is required")
		os.Exit(1)
	}

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		slog.Error("mirations folder not found", "path", migrationsPath)
		os.Exit(1)
	}

	dsnWithTable := appendDSNParam(storagePath, "x-migrations-table", migrationsTable)

	slog.Info("migrator starts...",
		"storage", storagePath,
		"migrations", migrationsPath,
		"table", migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsnWithTable,
	)
	if err != nil {
		slog.Error("migrator initialization was failed", "error", err)
		os.Exit(1)
	}

	m.Log = migrateLogger{}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no new migrations for apply")
			return
		}
		slog.Error("migrations applied error", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied successfully")
}

func SetupLogger(logFileName string, level slog.Level) {
	logDir := os.Getenv("LOG_PATH_MIGRATE")
	if logDir == "" {
		logDir = "logs"
	}

	path := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic("can not open log file: " + err.Error())
	}

	writer := io.MultiWriter(os.Stderr, logFile)

	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(t.Format("2006-01-02 15:04:05")),
				}
			}
			return a
		},
	}))

	slog.SetDefault(logger)

	slog.Info("Started logger",
		"file", path,
		"level", level.String(),
		"current time", time.Now().Format("15:04:05"))
}
