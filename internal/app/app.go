package app

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/service"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type App struct {
	DB      *sql.DB
	Service service.SongServiceInterface
	Logger  *slog.Logger
}

func NewApp(config *config.Config, logger *slog.Logger) (*App, error) {
	logger.Info("Initializing application", "config", config)
	dir, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get working directory", "error", err)
	} else {
		logger.Debug("Current working directory", "dir", dir)
	}
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
	logger.Debug("Connecting to database", "connection_string", psqlInfo)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Error("Failed to open database connection", "error", err)
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	logger.Info("Database connection opened successfully")

	err = db.Ping()
	if err != nil {
		logger.Error("Failed to ping database", "error", err)
		return nil, fmt.Errorf("ошибка проверки соединения: %v", err)
	}
	logger.Info("Database ping successful")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("Failed to create migration driver", "error", err)
		return nil, fmt.Errorf("ошибка создания драйвера миграций: %v", err)
	}
	logger.Debug("Migration driver created successfully")

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logger.Error("Failed to initialize migrations", "error", err)
		return nil, fmt.Errorf("ошибка инициализации миграций: %v", err)
	}
	logger.Debug("Migrations initialized successfully")

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error("Failed to apply migrations", "error", err)
		return nil, fmt.Errorf("ошибка применения миграций: %v", err)
	}
	logger.Info("Migrations applied successfully or no change")

	service := service.NewSongService(db, logger)
	return &App{
		DB:      db,
		Service: service,
		Logger:  logger,
	}, nil
}
