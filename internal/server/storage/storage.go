package storage

import (
	"context"
	"database/sql"
	"errors"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("update conflict")
)

// UserStorage is interface for adding and getting of user's data as login, password hash.
type UserStorage interface {
	// AddUser saves new user if login is unique or returns ErrAlreadyExists otherwise
	AddUser(ctx context.Context, login string, hashedPass string) error
	// GetUser returns userId and hashed password found by login or ErrNotFound if there is no such user
	GetUser(ctx context.Context, login string) (uID server.UserID, pwd string, err error)
	Close()
}

// DataStorage is interface for working with stored user's sensitive data
type DataStorage interface {
	GetByKey(ctx context.Context, userID server.UserID, key string) (*pb.Value, error)
	Put(ctx context.Context, userID server.UserID, key string, value *pb.Value) error
	GetAllKeysByUser(ctx context.Context, userID server.UserID) ([]string, error)
	GetAllByUser(ctx context.Context, userID server.UserID, version int32) (map[string]*pb.Value, int32, error)
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
