package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

//go:embed *.sql
var files embed.FS

func Run(db *sql.DB) error {
	goose.SetBaseFS(files)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	return goose.Up(db, ".")
}

// RunRemote aplica migrações pendentes diretamente no Turso Cloud.
// Necessário no modo sync: o Goose roda na réplica local, mas o DDL
// nem sempre é replicado ao remoto — apenas goose_db_version pode ser sincronizado.
func RunRemote(url, token string) error {
	if url == "" || token == "" {
		return nil
	}
	dbURL := fmt.Sprintf("%s?authToken=%s", url, token)
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		return fmt.Errorf("remote open: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("remote ping: %w", err)
	}
	return Run(db)
}
