package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthLocalService struct {
	db *sql.DB
}

func NewAuthLocalService(db *sql.DB) *AuthLocalService {
	return &AuthLocalService{db: db}
}

func (as *AuthLocalService) Register(ctx context.Context, email, password string) (int64, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return 0, fmt.Errorf("email and password are required")
	}
	if len(password) < 6 {
		return 0, fmt.Errorf("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	res, err := as.db.ExecContext(
		ctx,
		`INSERT INTO users(email, password_hash, currency_code, language, created_at) VALUES (?, ?, 'USD', 'en', ?)`,
		email,
		string(hash),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, fmt.Errorf("email already registered")
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read created user id: %w", err)
	}
	return userID, nil
}

func (as *AuthLocalService) Login(ctx context.Context, email, password string) (int64, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return 0, fmt.Errorf("email and password are required")
	}

	var userID int64
	var passwordHash string
	err := as.db.QueryRowContext(
		ctx,
		`SELECT id, password_hash FROM users WHERE email = ? LIMIT 1`,
		email,
	).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid email or password")
		}
		return 0, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return 0, fmt.Errorf("invalid email or password")
	}

	return userID, nil
}
