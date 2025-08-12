package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"realtime-docs/internal/app"
)

type Postgres struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewPostgres connects to postgres and returns a pool wrapper
func NewPostgres(ctx context.Context, cfg app.Config, log *slog.Logger) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, cfg.PGURL)
	if err != nil {
		return nil, err
	}
	return &Postgres{pool: pool, log: log}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

// CreateDoc inserts a new document owned by userID
func (p *Postgres) CreateDoc(ctx context.Context, title, userID string) (Doc, error) {
	row := p.pool.QueryRow(ctx, `
		INSERT INTO documents (title, bytes, version, created_by)
		VALUES ($1, ''::bytea, 0, $2)
		RETURNING id, title, bytes, version, created_by, created_at, updated_at
	`, title, userID)

	var d Doc
	if err := row.Scan(&d.ID, &d.Title, &d.Bytes, &d.Version, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return Doc{}, err
	}
	return d, nil
}

// ListDocs returns docs sorted by last update
func (p *Postgres) ListDocs(ctx context.Context, limit, offset int) ([]Doc, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, title, bytes, version, created_by, created_at, updated_at
		FROM documents
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Doc
	for rows.Next() {
		var d Doc
		if err := rows.Scan(&d.ID, &d.Title, &d.Bytes, &d.Version, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetDoc fetches a document by ID
func (p *Postgres) GetDoc(ctx context.Context, id string) (Doc, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, title, bytes, version, created_by, created_at, updated_at
		FROM documents
		WHERE id = $1
	`, id)

	var d Doc
	if err := row.Scan(&d.ID, &d.Title, &d.Bytes, &d.Version, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return Doc{}, err
	}
	return d, nil
}

// SaveDoc updates doc bytes, bumps version, and timestamp
func (p *Postgres) SaveDoc(ctx context.Context, id string, blob []byte) error {
	ct, err := p.pool.Exec(ctx, `
		UPDATE documents
		SET bytes = $2, version = version + 1, updated_at = NOW()
		WHERE id = $1
	`, id, blob)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("doc not found")
	}
	p.log.Info("doc.saved", "id", id, "bytes", len(blob))
	return nil
}