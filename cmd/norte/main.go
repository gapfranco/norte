package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"norte/config"
	"norte/internal/migrations"
	"norte/internal/storage"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

type application struct {
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
	db             *storage.TursoDB
	config         config.Config
	setupDone      atomic.Bool
	isDesktop      atomic.Bool
}

func main() {
	headless := flag.Bool("headless", false, "Roda como servidor HTTP sem janela")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := storage.NewTursoDB(storage.DBConfig{
		URL:       cfg.DBURL,
		Token:     cfg.DBToken,
		Mode:      cfg.DBMode,
		LocalPath: cfg.DBLocalPath,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Run(db.DB()); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if cfg.DBMode == "sync" {
		if err := migrations.RunRemote(cfg.DBURL, cfg.DBToken); err != nil {
			logger.Warn("failed to run remote migrations", "error", err)
		}
	}

	if err := db.Sync(context.Background()); err != nil {
		logger.Warn("db sync after migrations failed", "error", err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("failed to create template cache", "error", err)
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	app := &application{
		logger:         logger,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    form.NewDecoder(),
		db:             db,
		config:         cfg,
	}

	hasUsers, err := db.HasUsers()
	if err != nil {
		logger.Error("failed to check users", "error", err)
		os.Exit(1)
	}
	app.setupDone.Store(hasUsers)

	if *headless {
		runServer(app, cfg, logger)
	} else {
		runDesktop(app, cfg, logger)
	}
}

func runServer(app *application, cfg config.Config, logger *slog.Logger) {
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      app.routes(),
		ErrorLog:     log.New(os.Stderr, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://localhost%s", cfg.Addr)
	go openBrowser(url)

	logger.Info("starting server", "addr", cfg.Addr)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
