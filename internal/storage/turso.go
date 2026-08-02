package storage

import (
	"context"
	"database/sql"
	"fmt"

	"norte/internal/models"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"golang.org/x/crypto/bcrypt"
	turso "turso.tech/database/tursogo"
)

// DBConfig holds the database connection configuration.
type DBConfig struct {
	URL       string
	Token     string
	Mode      string // "local" | "remote" | "sync"
	LocalPath string // path to local SQLite file (used in "local" and "sync" modes)
}

type TursoDB struct {
	db     *sql.DB
	syncFn func(ctx context.Context) error
}

func enableForeignKeys(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	return err
}

// NewTursoDB creates a database connection according to cfg.Mode:
//   - "local"  — pure local SQLite, no network
//   - "sync"   — embedded replica: reads from local, syncs to/from Turso Cloud
//   - "remote" — (default) direct connection to Turso Cloud
func NewTursoDB(cfg DBConfig) (*TursoDB, error) {
	switch cfg.Mode {
	case "local":
		db, err := sql.Open("turso", cfg.LocalPath)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		return &TursoDB{db: db}, nil

	case "sync":
		ctx := context.Background()
		syncDb, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
			Path:      cfg.LocalPath,
			RemoteUrl: cfg.URL,
			AuthToken: cfg.Token,
		})
		if err != nil {
			return nil, fmt.Errorf("turso sync: %w", err)
		}
		db, err := syncDb.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("turso sync connect: %w", err)
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		syncFn := func(ctx context.Context) error {
			if err := syncDb.Push(ctx); err != nil {
				return fmt.Errorf("turso push: %w", err)
			}
			if _, err := syncDb.Pull(ctx); err != nil {
				return fmt.Errorf("turso pull: %w", err)
			}
			return nil
		}
		return &TursoDB{db: db, syncFn: syncFn}, nil

	default: // "remote"
		var dbURL string
		if cfg.Token != "" {
			dbURL = fmt.Sprintf("%s?authToken=%s", cfg.URL, cfg.Token)
		} else {
			dbURL = cfg.URL
		}
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		if err := enableForeignKeys(db); err != nil {
			return nil, err
		}
		return &TursoDB{db: db}, nil
	}
}

// Sync pulls remote changes and pushes local writes. No-op in "local" and "remote" modes.
func (t *TursoDB) Sync(ctx context.Context) error {
	if t.syncFn != nil {
		return t.syncFn(ctx)
	}
	return nil
}

func (t *TursoDB) DB() *sql.DB {
	return t.db
}

func (t *TursoDB) Close() error {
	return t.db.Close()
}

// ─── Users ────────────────────────────────────────────────────────────────────

func (t *TursoDB) HasUsers() (bool, error) {
	var count int
	err := t.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0, err
}

func (t *TursoDB) CreateUser(usuario, nome, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = t.db.Exec("INSERT INTO users (usuario, nome, senha) VALUES (?, ?, ?)", usuario, nome, string(hash))
	return err
}

func (t *TursoDB) Authenticate(usuario, senha string) (*models.User, error) {
	var user models.User
	var hash string
	err := t.db.QueryRow(
		"SELECT id, usuario, nome, senha FROM users WHERE lower(usuario) = lower(?)",
		usuario,
	).Scan(&user.ID, &user.Usuario, &user.Nome, &hash)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)); err != nil {
		return nil, fmt.Errorf("senha inválida")
	}
	return &user, nil
}

func (t *TursoDB) ListUsers(limit, offset int) ([]models.User, int, error) {
	return t.ListUsersFilter("", "", limit, offset)
}

func (t *TursoDB) ListUsersFilter(usuario, nome string, limit, offset int) ([]models.User, int, error) {
	like := func(s string) string { return "%" + s + "%" }
	var total int
	if err := t.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE usuario LIKE ? AND nome LIKE ?",
		like(usuario), like(nome),
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := t.db.Query(
		"SELECT id, usuario, nome FROM users WHERE usuario LIKE ? AND nome LIKE ? ORDER BY usuario LIMIT ? OFFSET ?",
		like(usuario), like(nome), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Usuario, &u.Nome); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (t *TursoDB) GetUser(id int) (*models.User, error) {
	var u models.User
	err := t.db.QueryRow("SELECT id, usuario, nome FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Usuario, &u.Nome)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (t *TursoDB) GetUserByUsuario(usuario string) (*models.User, error) {
	var u models.User
	err := t.db.QueryRow(
		"SELECT id, usuario, nome FROM users WHERE lower(usuario) = lower(?)",
		usuario,
	).Scan(&u.ID, &u.Usuario, &u.Nome)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (t *TursoDB) UpdateUserNome(id int, nome string) error {
	_, err := t.db.Exec("UPDATE users SET nome = ? WHERE id = ?", nome, id)
	return err
}

func (t *TursoDB) CheckUserPassword(id int, senha string) error {
	var hash string
	err := t.db.QueryRow("SELECT senha FROM users WHERE id = ?", id).Scan(&hash)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)); err != nil {
		return fmt.Errorf("senha inválida")
	}
	return nil
}

func (t *TursoDB) UpdateUserPassword(id int, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = t.db.Exec("UPDATE users SET senha = ? WHERE id = ?", string(hash), id)
	return err
}

func (t *TursoDB) DeleteUser(id int) error {
	_, err := t.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
