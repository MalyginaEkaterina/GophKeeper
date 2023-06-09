package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server/storage"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUnauthorized      = errors.New("unauthorized")
)

const (
	tokenExpiry = 10 * time.Minute
)

// AuthService is interface for registration and authorization
type AuthService interface {
	RegisterUser(ctx context.Context, login string, pass string) error
	AuthUser(ctx context.Context, login string, pass string) (server.Token, error)
	CheckToken(token string) (server.UserID, error)
}

type AuthServiceImpl struct {
	Store     storage.UserStorage
	SecretKey []byte
}

// RegisterUser generates hash of password and saves new user with this hash in hex format
func (a *AuthServiceImpl) RegisterUser(ctx context.Context, login string, pass string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("create password hash error: %w", err)
	}
	return a.Store.AddUser(ctx, login, hex.EncodeToString(hashedPass))
}

// AuthUser gets user's data by login, checks the password and returns token if the password is correct or
// ErrIncorrectPassword otherwise
func (a *AuthServiceImpl) AuthUser(ctx context.Context, login string, pass string) (server.Token, error) {
	userID, hashedPass, err := a.Store.GetUser(ctx, login)
	if err != nil {
		return "", err
	}
	hash, err := hex.DecodeString(hashedPass)
	if err != nil {
		return "", fmt.Errorf("decode hashed password error: %w", err)
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(pass))
	if err != nil {
		return "", ErrIncorrectPassword
	}
	token, err := a.CreateToken(userID)
	if err != nil {
		return "", fmt.Errorf("create token error: %w", err)
	}
	return token, nil
}

// CreateToken takes userId and expiry time, creates sign of this data with SecretKey and appends it to this data.
// Returns the token from this data in hex
func (a *AuthServiceImpl) CreateToken(id server.UserID) (server.Token, error) {
	data := binary.BigEndian.AppendUint32(nil, uint32(id))
	expTime := time.Now().Add(tokenExpiry).Unix()
	data = binary.BigEndian.AppendUint64(data, uint64(expTime))
	h := hmac.New(sha256.New, a.SecretKey)
	h.Write(data)
	sign := h.Sum(nil)
	data = append(data, sign...)
	return server.Token(hex.EncodeToString(data)), nil
}

// CheckToken takes userId and expiry time from token, checks if token was not expired and checks the sign in this token.
// Returns userId
func (a *AuthServiceImpl) CheckToken(token string) (server.UserID, error) {
	data, err := hex.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("decode token error: %w", err)
	}
	if len(data) < 12 {
		return 0, ErrUnauthorized
	}
	userID := binary.BigEndian.Uint32(data[:4])
	expTime := binary.BigEndian.Uint64(data[4:12])
	if time.Now().Unix() > int64(expTime) {
		return 0, ErrUnauthorized
	}
	h := hmac.New(sha256.New, a.SecretKey)
	h.Write(data[:12])
	sign := h.Sum(nil)
	if !hmac.Equal(sign, data[12:]) {
		return 0, ErrUnauthorized
	}
	return server.UserID(userID), nil
}

var _ AuthService = (*AuthServiceImpl)(nil)
