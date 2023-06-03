package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
)

type DBDataStorage struct {
	db            *sql.DB
	getByKey      *sql.Stmt
	insert        *sql.Stmt
	update        *sql.Stmt
	getKeysByUser *sql.Stmt
	getAllByUser  *sql.Stmt
}

var _ DataStorage = (*DBDataStorage)(nil)

func NewDBDataStorage(db *sql.DB) (*DBDataStorage, error) {
	stmtGetByKey, err := db.Prepare("SELECT data, metadata, version from data WHERE user_id = $1 AND key = $2")
	if err != nil {
		return nil, err
	}
	stmtInsert, err := db.Prepare("INSERT INTO data (user_id, key, data, metadata, version) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING RETURNING key")
	if err != nil {
		return nil, err
	}
	stmtUpdate, err := db.Prepare("UPDATE data SET data = $3, metadata = $4, version = $5 WHERE user_id = $1 AND key = $2 AND version < $5")
	if err != nil {
		return nil, err
	}
	stmtGetKeysByUser, err := db.Prepare("SELECT key from data WHERE user_id = $1")
	if err != nil {
		return nil, err
	}
	stmtGetAllByUser, err := db.Prepare("SELECT key, data, metadata, version from data WHERE user_id = $1")
	if err != nil {
		return nil, err
	}

	return &DBDataStorage{
		db:            db,
		getByKey:      stmtGetByKey,
		insert:        stmtInsert,
		update:        stmtUpdate,
		getKeysByUser: stmtGetKeysByUser,
		getAllByUser:  stmtGetAllByUser,
	}, nil
}

func (d *DBDataStorage) GetByKey(ctx context.Context, userID server.UserID, key string) (*pb.Value, error) {
	row := d.getByKey.QueryRowContext(ctx, userID, key)
	var result pb.Value
	err := row.Scan(&result.Data, &result.Metadata, &result.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("get data by key error: %w", err)
	}
	return &result, nil
}

func (d *DBDataStorage) Put(ctx context.Context, userID server.UserID, key string, value *pb.Value) error {
	if value.Version == 1 {
		row := d.insert.QueryRowContext(ctx, userID, key, value.Data, value.Metadata, value.Version)
		var insKey string
		err := row.Scan(&insKey)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrConflict
		} else if err != nil {
			return fmt.Errorf("insert data error: %w", err)
		}
		return nil
	} else {
		res, err := d.update.ExecContext(ctx, userID, key, value.Data, value.Metadata, value.Version)
		if err != nil {
			return fmt.Errorf("update data error: %w", err)
		}
		count, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("get result of update data error: %w", err)
		}
		if count == 0 {
			return ErrConflict
		}
		return nil
	}
}

func (d *DBDataStorage) GetAllKeysByUser(ctx context.Context, userID server.UserID) ([]string, error) {
	rows, err := d.getKeysByUser.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userKeys []string

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		userKeys = append(userKeys, key)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userKeys, nil
}

func (d *DBDataStorage) GetAllByUser(ctx context.Context, userID server.UserID) (map[string]*pb.Value, error) {
	rows, err := d.getAllByUser.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userValues := make(map[string]*pb.Value)

	for rows.Next() {
		var key string
		var value pb.Value
		err = rows.Scan(&key, &value.Data, &value.Metadata, &value.Version)
		if err != nil {
			return nil, err
		}
		userValues[key] = &value
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userValues, nil
}

func (d *DBDataStorage) Close() {
	d.getByKey.Close()
	d.insert.Close()
	d.update.Close()
	d.getKeysByUser.Close()
	d.getAllByUser.Close()
}
