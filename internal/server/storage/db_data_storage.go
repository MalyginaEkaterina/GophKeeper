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
	db             *sql.DB
	getByKey       *sql.Stmt
	insert         *sql.Stmt
	update         *sql.Stmt
	getKeysByUser  *sql.Stmt
	getAllByUser   *sql.Stmt
	updDataVersion *sql.Stmt
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
	const selectAll = `
		SELECT d.key, d.data, d.metadata, d.version, u.data_version
		FROM data d JOIN users u on u.id = d.user_id
		WHERE u.id = $1 AND u.data_version > $2
	`
	stmtGetAllByUser, err := db.Prepare(selectAll)
	if err != nil {
		return nil, err
	}
	stmtUpdDataVersion, err := db.Prepare("UPDATE users SET data_version = data_version+1 WHERE id = $1")
	if err != nil {
		return nil, err
	}

	return &DBDataStorage{
		db:             db,
		getByKey:       stmtGetByKey,
		insert:         stmtInsert,
		update:         stmtUpdate,
		getKeysByUser:  stmtGetKeysByUser,
		getAllByUser:   stmtGetAllByUser,
		updDataVersion: stmtUpdDataVersion,
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
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}
	defer tx.Rollback()
	if value.Version == 1 {
		txInsertStmt := tx.StmtContext(ctx, d.insert)
		row := txInsertStmt.QueryRowContext(ctx, userID, key, value.Data, value.Metadata, value.Version)
		var insKey string
		err := row.Scan(&insKey)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrConflict
		} else if err != nil {
			return fmt.Errorf("insert data error: %w", err)
		}
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
	}
	txUpdDataVersionStmt := tx.StmtContext(ctx, d.updDataVersion)
	_, err = txUpdDataVersionStmt.ExecContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("update data version error: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	return nil
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

func (d *DBDataStorage) GetAllByUser(ctx context.Context, userID server.UserID, version int32) (map[string]*pb.Value, int32, error) {
	rows, err := d.getAllByUser.QueryContext(ctx, userID, version)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	userValues := make(map[string]*pb.Value)
	var currVersion int32

	for rows.Next() {
		var key string
		var value pb.Value
		err = rows.Scan(&key, &value.Data, &value.Metadata, &value.Version, &currVersion)
		if err != nil {
			return nil, 0, err
		}
		userValues[key] = &value
	}

	err = rows.Err()
	if err != nil {
		return nil, 0, err
	}

	if len(userValues) == 0 {
		return nil, 0, ErrNotFound
	}

	return userValues, currVersion, nil
}

func (d *DBDataStorage) Close() {
	d.getByKey.Close()
	d.insert.Close()
	d.update.Close()
	d.getKeysByUser.Close()
	d.getAllByUser.Close()
	d.updDataVersion.Close()
}
