package grpcapp

import (
	"fmt"
	"github.com/AramLab/AuthService/internal/server/grpc/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// Пишем функции для создания приложения, ее запуска и так далее,
// данные функции будут вызываться в файле app/app.go

type App struct {
	log        *zap.SugaredLogger
	gRPCServer *grpc.Server
	port       int
}

func NewGRPCApp(log *zap.SugaredLogger, authService auth.Auth, port int) *App {
	gRPCServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Infof("grpc server starts on port: %s", l.Addr().String())

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	a.log.Infof("stopping gRPC server")
	a.gRPCServer.GracefulStop()
}
