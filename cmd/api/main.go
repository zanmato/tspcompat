package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"tsp/internal/sign"

	"github.com/BurntSushi/toml"
	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron/v3"
	"github.com/z0ne-dev/mgx/v2"
	"go.uber.org/zap"
)

func main() {
	// Load config
	cfg := struct {
		App struct {
			Listen string
			Local  bool
		}
		Signs struct {
			DataURL         string `toml:"data_url"`
			RefreshSchedule string `toml:"refresh_schedule"`
		}
		Database struct {
			DSN string
		}
	}{}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to get working directory: %v", err)
	}

	if _, err := toml.DecodeFile(path.Join(wd, "config.toml"), &cfg); err != nil {
		log.Fatalf("unable to decode config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("unable to initialize zap logger: %v", err)
	}

	// Open DB
	connConf, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		logger.Fatal("unable to parse database DSN", zap.Error(err))
	}

	connConf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   zapadapter.NewLogger(logger),
		LogLevel: tracelog.LogLevelWarn,
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), connConf)
	if err != nil {
		logger.Fatal("unable to connect to database", zap.Error(err))
	}
	defer dbpool.Close()

	migrator, err := mgx.New(
		mgx.Migrations(
			mgx.NewMigration("schema", migrateSchema),
			mgx.NewMigration("views", migrateViews),
		),
		mgx.Log(&migrationLogger{Logger: logger}),
	)
	if err != nil {
		logger.Fatal("unable to create migrator", zap.Error(err))
	}

	if err := migrator.Migrate(context.Background(), dbpool); err != nil {
		logger.Fatal("unable to migrate database", zap.Error(err))
	}

	// Create a sync client
	syncClient, err := sign.NewSyncClient(dbpool)
	if err != nil {
		logger.Fatal("unable to create sync client", zap.Error(err))
	}

	// Check if we have data, otherwise sync immediately
	var hasData bool
	if err := dbpool.QueryRow(
		context.Background(),
		`SELECT EXISTS (SELECT * FROM signs)`,
	).Scan(&hasData); err != nil {
		logger.Fatal("unable to check if there is data", zap.Error(err))
	}

	if !hasData {
		logger.Info("no data found, running sync job")
		if err := syncData(dbpool, syncClient, cfg.Signs.DataURL); err != nil {
			logger.Fatal("unable to sync data", zap.Error(err))
		}
	}

	// Schedule data sync
	c := cron.New(cron.WithLocation(time.Local))
	if _, err := c.AddFunc(cfg.Signs.RefreshSchedule, syncJob(dbpool, syncClient, cfg.Signs.DataURL, logger)); err != nil {
		logger.Fatal("unable to add sync job", zap.Error(err))
	}

	c.Start()

	// HTTP Router
	router := httprouter.New()

	if cfg.App.Local {
		router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", "*")
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			w.WriteHeader(http.StatusNoContent)
		})
	} else {
		// Front controller
		router.ServeFiles("/assets/*filepath", http.Dir("/app/dist/assets"))
		router.ServeFiles("/static/*filepath", http.Dir("/app/dist/static"))
		router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/app/dist/index.html")
		})
	}

	// Create signs API controller
	signsController := sign.NewAPIController(dbpool, logger)
	router.GET("/api/words", localCORS(cfg.App.Local, signsController.WordIndex))
	router.GET("/api/categories", localCORS(cfg.App.Local, signsController.CategoryIndex))
	router.GET("/api/signs/:id", localCORS(cfg.App.Local, signsController.SignShow))

	srv := http.Server{
		Addr:    cfg.App.Listen,
		Handler: router,
	}

	// Gracefully shutdown the server
	connClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
		}
		cronCtx := c.Stop()
		<-cronCtx.Done()
		close(connClosed)
	}()

	logger.Info("starting server", zap.String("listen", srv.Addr))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("server quit unexpectedly", zap.Error(err))
	}

	<-connClosed
}

func syncJob(db *pgxpool.Pool, sc *sign.SyncClient, dataURL string, logger *zap.Logger) func() {
	return func() {
		started := time.Now()
		logger.Info("starting sync job")
		if err := syncData(db, sc, dataURL); err != nil {
			logger.Error("failed loading data", zap.Error(err))
		}
		logger.Info("sync job finished", zap.Duration("duration", time.Since(started)))
	}
}

func syncData(db *pgxpool.Pool, sc *sign.SyncClient, dataURL string) error {
	if err := sc.Sync(context.Background(), dataURL); err != nil {
		return fmt.Errorf("failed loading data from %s: %w", dataURL, err)
	}

	if _, err := db.Exec(
		context.Background(),
		`REFRESH MATERIALIZED VIEW CONCURRENTLY signs_view`,
	); err != nil {
		return fmt.Errorf("unable to refresh materialized view: %w", err)
	}

	if _, err := db.Exec(
		context.Background(),
		`REFRESH MATERIALIZED VIEW CONCURRENTLY words_view`,
	); err != nil {
		return fmt.Errorf("unable to refresh materialized view: %w", err)
	}

	if _, err := db.Exec(
		context.Background(),
		`REFRESH MATERIALIZED VIEW CONCURRENTLY categories_view`,
	); err != nil {
		return fmt.Errorf("unable to refresh materialized view: %w", err)
	}

	return nil
}

func localCORS(local bool, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if local {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", "*")
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		}
		next(w, r, ps)
	}
}
