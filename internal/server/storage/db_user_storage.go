package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
)

type DBUserStorage struct {
	db             *sql.DB
	insertUser     *sql.Stmt
	getUserByLogin *sql.Stmt
}

var _ UserStorage = (*DBUserStorage)(nil)

func NewDBUserStorage(db *sql.DB) (*DBUserStorage, error) {
	stmtInsertUser, err := db.Prepare("INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id")
	if err != nil {
		return nil, err
	}
	stmtGetUserByLogin, err := db.Prepare("SELECT id, password from users WHERE login = $1")
	if err != nil {
		return nil, err
	}

	return &DBUserStorage{
		db:             db,
		insertUser:     stmtInsertUser,
		getUserByLogin: stmtGetUserByLogin,
	}, nil
}

func (d *DBUserStorage) Close() {
	d.insertUser.Close()
	d.getUserByLogin.Close()
}

func (d *DBUserStorage) AddUser(ctx context.Context, login string, hashedPass string) error {
	row := d.insertUser.QueryRowContext(ctx, login, hashedPass)
	var id server.UserID
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAlreadyExists
	} else if err != nil {
		return fmt.Errorf("insert user error: %w", err)
	}
	return nil
}

func (d *DBUserStorage) GetUser(ctx context.Context, login string) (server.UserID, string, error) {
	row := d.getUserByLogin.QueryRowContext(ctx, login)
	var id server.UserID
	var pass string
	err := row.Scan(&id, &pass)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	} else if err != nil {
		return 0, "", fmt.Errorf("get user error: %w", err)
	}
	return id, pass, nil
}
