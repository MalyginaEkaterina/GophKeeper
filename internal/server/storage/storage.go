package storage

import (
	"context"
	"database/sql"
	"errors"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("update conflict")
)

type UserStorage interface {
	AddUser(ctx context.Context, login string, hashedPass string) error
	GetUser(ctx context.Context, login string) (uID server.UserID, pwd string, err error)
	Close()
}

type DataStorage interface {
	GetByKey(ctx context.Context, userID server.UserID, key string) (*pb.Value, error)
	Put(ctx context.Context, userID server.UserID, key string, value *pb.Value) error
	GetAllKeysByUser(ctx context.Context, userID server.UserID) ([]string, error)
	GetAllByUser(ctx context.Context, userID server.UserID) (map[string]*pb.Value, error)
}

func DoMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
