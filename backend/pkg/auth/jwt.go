package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey int

const userKey ctxKey = 1

// WithUser adds a user ID to the context
func WithUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

// UserID extracts the user ID from the context, defaults to "anon"
func UserID(ctx context.Context) string {
	v := ctx.Value(userKey)
	if v == nil {
		return "anon"
	}
	return v.(string)
}

// JWT wraps a signing secret for issuing/verifying tokens
type JWT struct{ secret []byte }

// New creates a new JWT signer/verifier.
func New(secret string) *JWT { return &JWT{secret: []byte(secret)} }

// Verify checks a token and returns the sub (user ID) claim
func (j *JWT) Verify(tok string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tok, claims, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return "", err
	}
	uid, _ := claims["sub"].(string)
	if uid == "" {
		return "", errors.New("no sub")
	}
	return uid, nil
}

// Sign creates a token for uid with the given TTL
func (j *JWT) Sign(uid string, ttl time.Duration) (string, error) {
	if uid == "" {
		return "", errors.New("empty uid")
	}
	claims := jwt.MapClaims{
		"sub": uid,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(ttl).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.secret)
}
