package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

// normEmail trims and lowercases the email (needed if DB col isnt citext)
func normEmail(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// CreateUser inserts a new user with a hashed password
func (p *Postgres) CreateUser(ctx context.Context, email, password string) (User, error) {
	email = normEmail(email)
	if email == "" || password == "" {
		return User{}, errors.New("missing email or password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	row := p.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, email, string(hash))

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
		return User{}, err
	}
	return u, nil
}

// GetUserByEmail returns the user + hashed password for login verification
func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	email = normEmail(email)

	row := p.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email)

	var u User
	var hash string
	if err := row.Scan(&u.ID, &u.Email, &hash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", errors.New("not found")
		}
		return User{}, "", err
	}
	return u, hash, nil
}

// VerifyUser checks email + password match
func (p *Postgres) VerifyUser(ctx context.Context, email, password string) (User, error) {
	u, hash, err := p.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return User{}, errors.New("invalid credentials")
	}

	return u, nil
}
