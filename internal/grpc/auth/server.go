package auth

import (
	"context"
	"errors"
	repo "github.com/AramLab/AuthService/internal/repository/postgres"
	"github.com/AramLab/AuthService/internal/service"
	pb "github.com/AramLab/protos/gen/go/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// реализовать данный интерфейс в services/auth
type Auth interface {
	RegisterNewUser(ctx context.Context, username string, password string, email string, firstName string, lastName string) (*uuid.UUID, error)
	Login(ctx context.Context, username, password string) (token string, err error)
}

// Комментарий для меня: server = handler(слой)
type server struct {
	pb.UnimplementedAuthServiceServer
	auth Auth
}

func RegisterAuthServiceServer(gRPCServer *grpc.Server, auth Auth) {
	pb.RegisterAuthServiceServer(gRPCServer, &server{auth: auth})
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.Username, req.Password, req.Email, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, repo.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}
	return &pb.RegisterResponse{UserID: userID.String()}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	token, err := s.auth.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid username or password")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &pb.LoginResponse{Token: token}, nil
}
