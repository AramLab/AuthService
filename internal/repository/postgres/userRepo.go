package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/AramLab/AuthService/internal/config"
	"github.com/AramLab/AuthService/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

const (
	checkUserByUserNameQuery      = `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1);`
	saveUserQuery                 = `INSERT INTO users (username, password_hash, email, first_name, last_name) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	getUserByUsernameQuery        = `SELECT id, username, password_hash, email, first_name, last_name, is_active, role FROM users WHERE username = $1;`
	setLastLoginAtByUsernameQuery = `UPDATE users SET last_login_at = now() WHERE username = $1;`
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(ctx context.Context, cfg config.PostgresCfg) (*UserRepo, error) {
	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String())
	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create poool: %w", err)
	}

	return &UserRepo{pool: pool}, nil
}

func (u *UserRepo) SaveUser(ctx context.Context, username string, passHash []byte, email, firstName, lastName string) (userID *uuid.UUID, err error) {
	const op = "repository.postgres.userRepo.go"

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}

	defer func() {
		if err != nil {
			// если в err что-то записано, откатываем
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("%s: rollback error: %v (original error: %w)", op, rbErr, err)
			}
		} else {
			// если ошибок нет, коммитим
			err = tx.Commit(ctx)
		}
	}()

	var exists bool
	err = tx.QueryRow(ctx, checkUserByUserNameQuery, username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to check existing user: %w", op, err)
	}

	if exists {
		return nil, fmt.Errorf("%s: %w", op, ErrUserExists)
	}

	var id uuid.UUID
	err = tx.QueryRow(ctx, saveUserQuery, username, passHash, email, firstName, lastName).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save user: %w", op, err)
	}

	return &id, nil
}

func (u *UserRepo) GetUser(ctx context.Context, username string) (*models.User, error) {
	const op = "repository.postgres.userRepo.GetUser"

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	row := tx.QueryRow(ctx, getUserByUsernameQuery, username)

	var user models.User
	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound // возвращаем свою ошибку, чтобы сервис мог проверить
		}
		return nil, fmt.Errorf("%s: failed to scan user: %w", op, err)
	}

	_, err = tx.Exec(ctx, setLastLoginAtByUsernameQuery, username)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to update last_login_at: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	tx = nil

	return &user, nil
}
