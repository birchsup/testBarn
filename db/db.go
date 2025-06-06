package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"time"
)

var DBPool *pgxpool.Pool

func InitDB() {
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v\n", err)
	}

	// Устанавливаем логгер и уровень логирования
	config.ConnConfig.Logger = &customLogger{}
	config.ConnConfig.LogLevel = pgx.LogLevelDebug

	DBPool, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Database connection initialized successfully")
}

// customLogger implements pgx.Logger
type customLogger struct{}

func (l *customLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	if level >= pgx.LogLevelDebug {
		log.Printf("[%s] [%s] %s\n", time.Now().Format(time.RFC3339Nano), level, msg)
		if data != nil {
			log.Printf("Детали запроса: %+v\n", data)
		}
	}
}
