package main

import (
	"context"
	config "golang-test-task/internal/configs"
	"golang-test-task/internal/handlers"
	postgreSQL "golang-test-task/internal/repository/postgres"
	"golang-test-task/internal/service"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	sl "github.com/gladinov/mylogger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func setupServer(logg *slog.Logger, repo *postgreSQL.Storage) *echo.Echo {
	svc := service.New(logg, repo)
	handlers := handlers.NewHandlers(logg, svc)

	router := echo.New()
	router.Use(middleware.CORS())
	router.Use(handlers.ContextHeaderTraceIdMiddleWare)
	router.Use(handlers.LoggerMiddleWare)

	router.POST("/numbers", handlers.SaveNumberAndGetSortedNumbers)

	return router
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conf := config.MustInitConfig()

	logg := sl.NewLogger(conf.Env)

	logg.Info("start app",
		slog.String("env", conf.Env))

	postgresHost, err := conf.PostgresHost.GetStringHost()
	if err != nil {
		logg.Error("could not get postgres host from config")
		return
	}
	creator := postgreSQL.NewPoolCreator(postgresHost)
	repo := postgreSQL.MustInitNewStorageWithCreator(ctx, logg, creator)

	defer func() {
		if repo != nil {
			repo.CloseDB()
		}
	}()

	e := setupServer(logg, repo)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			logg.Error("Failed to shutdown server:", slog.String("error", err.Error()))
		}
	}()

	address := conf.Clients.TestAppClient.GetTestAppClientAddress()
	logg.Info("run test App", slog.String("address", address))
	e.Start(address)
}
