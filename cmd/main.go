package main

import (
	"context"
	"github.com/AramLab/AuthService/internal/app"
	"github.com/AramLab/AuthService/internal/config"
	"github.com/AramLab/AuthService/pkg/jwt"
	"github.com/AramLab/AuthService/pkg/logger"
	"github.com/AramLab/AuthService/pkg/migration"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	if err := migration.RunMigrations(cfg.PostgresCfg); err != nil {
		appLogger.Fatal(errors.Wrap(err, "failed to run migration"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenManager := jwt.NewJwtManager([]byte(cfg.Secret))

	application := app.New(ctx, appLogger, cfg.GrpcCfg.Port, cfg.PostgresCfg, tokenManager)
	if application == nil {
		log.Fatal("failed to initialize app")
	}

	go func() {
		appLogger.Info("starting server")
		application.GRPCServer.MustRun()
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	cancel()
	application.GRPCServer.Stop()

	appLogger.Info("Shutting down gracefully...")
}
