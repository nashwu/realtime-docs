package store

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"log/slog"
)

// Embed files from a subfolder next to this file
//go:embed migrations/*.sql
var migrations embed.FS


// RunMigrations executes all embedded .sql files in order
func RunMigrations(ctx context.Context, p *Postgres, log *slog.Logger) error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := p.pool.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		log.Info("migration.applied", "file", e.Name())
	}
	return nil
}
