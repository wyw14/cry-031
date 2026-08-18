package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wyw/cry031-volunteer/internal/application"
	"github.com/wyw/cry031-volunteer/internal/config"
	"github.com/wyw/cry031-volunteer/internal/domain"
	"github.com/wyw/cry031-volunteer/internal/platform"
	"github.com/wyw/cry031-volunteer/internal/repository"
	"github.com/wyw/cry031-volunteer/internal/service"
	httptransport "github.com/wyw/cry031-volunteer/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger, err := platform.NewLogger(os.Getenv("APP_ENV"))
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	clock := service.RealClock{}
	initial := domain.NewState()
	if cfg.SeedDemo {
		initial = repository.DemoState(clock.Now())
	}
	var store application.Store
	closeStore := func() {}
	var ready atomic.Bool
	ready.Store(true)
	if cfg.StoreMode == "postgres" {
		if cfg.DatabaseURL == "" {
			logger.Fatal("DATABASE_URL is required when STORE_MODE=postgres")
		}
		postgresStore, err := repository.NewPostgresStore(context.Background(), cfg.DatabaseURL, initial)
		if err != nil {
			logger.Fatal("initialize postgres store", zap.Error(err))
		}
		store = postgresStore
		closeStore = postgresStore.Close
	} else {
		store = repository.NewMemoryStore(initial)
	}
	defer closeStore()

	notifier := &service.LocalNotifier{}
	engine := application.NewEngine(store, clock, notifier)
	attachments := service.LocalFileStore{Root: cfg.AttachmentDir, MaxBytes: cfg.MaxUploadBytes}
	router := httptransport.NewRouter(engine, attachments, logger, cfg.RequestTimeout, ready.Load)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("volunteer service started", zap.String("address", cfg.HTTPAddr), zap.String("store", cfg.StoreMode))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	ready.Store(false)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}
