package app

import (
	"context"
	grpcapp "github.com/AramLab/AuthService/internal/app/grpc"
	"github.com/AramLab/AuthService/internal/config"
	ps "github.com/AramLab/AuthService/internal/repository/postgres"
	svc "github.com/AramLab/AuthService/internal/service"
	"go.uber.org/zap"
)

// Пишем скрипт(New) создания программы, чтобы разгрузить main.go

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, log *zap.SugaredLogger, grpcPort int, cfg config.PostgresCfg) *App {
	// инициализируем postgres(repository)
	repository, err := ps.NewUserRepo(ctx, cfg)
	if err != nil {
		log.Errorf("failed to create repository instance: %w", err)
	}

	// инициализируем сервис(services)
	authService := svc.NewAuthService(log, repository, repository)

	grpcApp := grpcapp.NewGRPCApp(log, authService, grpcPort)
	return &App{GRPCServer: grpcApp}
}
