package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"pvz/internal/app"
)

func main() {
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")
	serverPort := os.Getenv("SERVER_PORT")
	secret := os.Getenv("SECRET")
	grpcPort := os.Getenv("GRPC_PORT")
	prometheusPort := os.Getenv("PROMETHEUS_PORT")

	levelStr, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		logrus.Error("Строка уровня логгирования не установлена")
		levelStr = "info"
	}
	level, err := logrus.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		logrus.WithError(err).Error("Ошибка")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.Infof("Установлен уровень логирования: %s", level.String())

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" || serverPort == "" || secret == "" || grpcPort == "" || prometheusPort == "" {
		log.Fatal("Не все переменные окружения заданы")
	}

	jwtSecret := []byte(secret)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := pgxpool.New(context.Background(), dsn)

	if err != nil {
		logrus.WithError(err).Fatal("Ошибка подключения к БД")
	}

	application := app.NewApp(db, serverPort, grpcPort, prometheusPort, jwtSecret)
	if err := application.Run(); err != nil {
		db.Close()
		logrus.WithError(err).Fatal("Ошибка выполнения приложения")
	}
}
