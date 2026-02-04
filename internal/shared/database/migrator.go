package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/goku-m/main/internal/shared/config"

	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/**/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	return MigrateAll(ctx, logger, cfg)
}

func MigrateAll(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// URL-encode the password
	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	baseTree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}

	projectDirs, err := discoverProjectDirs(baseTree)
	if err != nil {
		return err
	}

	if len(projectDirs) == 0 {
		if err := migrateSubtree(ctx, logger, conn, baseTree, ".", "schema_version"); err != nil {
			return err
		}
		return nil
	}

	logger.Info().Strs("projects", projectDirs).Msg("discovered migration project folders")
	for _, project := range projectDirs {
		schemaTable := fmt.Sprintf("schema_version_%s", strings.ToLower(project))
		logger.Info().Str("project", project).Str("schema_table", schemaTable).Msg("running migrations")
		if err := migrateSubtree(ctx, logger, conn, baseTree, project, schemaTable); err != nil {
			return err
		}
	}

	return nil
}

func discoverProjectDirs(baseTree fs.FS) ([]string, error) {
	// Prefer directory listing when available.
	entries, err := fs.ReadDir(baseTree, ".")
	if err == nil {
		var dirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				dirs = append(dirs, entry.Name())
			}
		}
		if len(dirs) > 0 {
			return dirs, nil
		}
	}

	// Fallback: infer directory names by walking files (works with embed FS).
	dirSet := make(map[string]struct{})
	walkErr := fs.WalkDir(baseTree, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		parts := strings.Split(path, "/")
		if len(parts) > 1 {
			dirSet[parts[0]] = struct{}{}
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walking migrations: %w", walkErr)
	}
	var dirs []string
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

func migrateSubtree(
	ctx context.Context,
	logger *zerolog.Logger,
	conn *pgx.Conn,
	baseTree fs.FS,
	dir string,
	schemaTable string,
) error {
	var subtree fs.FS
	var err error
	if dir == "." {
		subtree = baseTree
	} else {
		subtree, err = fs.Sub(baseTree, dir)
		if err != nil {
			return fmt.Errorf("retrieving database migrations subtree %q: %w", dir, err)
		}
	}

	m, err := tern.NewMigrator(ctx, conn, schemaTable)
	if err != nil {
		return fmt.Errorf("constructing database migrator for %q: %w", dir, err)
	}
	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations for %q: %w", dir, err)
	}
	if len(m.Migrations) == 0 {
		return nil
	}
	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retreiving current database migration version for %q", dir)
	}
	if err := m.Migrate(ctx); err != nil {
		return err
	}
	if from == int32(len(m.Migrations)) {
		logger.Info().Msgf("database schema up to date for %s, version %d", schemaTable, len(m.Migrations))
	} else {
		logger.Info().Msgf("migrated database schema for %s, from %d to %d", schemaTable, from, len(m.Migrations))
	}
	return nil
}
