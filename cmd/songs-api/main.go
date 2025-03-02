package main

import (
	_ "awesomeProject/docs" // Импорт сгенерированной документации
	"awesomeProject/internal/app"
	"awesomeProject/internal/config"
	"awesomeProject/internal/handler"
	"awesomeProject/internal/logger"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Songs API
// @version 1.0
// @description Это API для управления песнями и их текстами
// @host localhost:1323
// @BasePath /
func main() {
	logger := logger.NewLogger()
	logger.Info("Starting...")
	config := config.NewConfig(logger)
	appInstance, err := app.NewApp(config, logger)
	if err != nil {
		logger.Error("Ошибка инициализации приложения", "error", err)
		fmt.Println(err)
		return
	}
	defer appInstance.DB.Close()
	h := handler.NewHandler(appInstance.Service, logger)
	e := echo.New()
	logger.Debug("Registering routes")
	e.GET("/songs", h.GetHandler)
	e.POST("/songs", h.PostHandler)
	e.DELETE("/songs", h.DeleteHandler)
	e.PATCH("/songs", h.PatchHandler)
	e.GET("/songs/:id/verses", h.GetVersesHandler)
	e.GET("/songs/verses/search", h.SearchVersesHandler)
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	logger.Info("Server starting on :1323    port")
	if err := e.Start(":1323"); err != nil {
		logger.Error("Server falied to start", "error", err)
		e.Logger.Fatal(err)
	}
}
