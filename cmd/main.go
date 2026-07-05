// Package main is the entrypoint for the DevHelp backend.
//
// It wires all dependencies together using explicit constructor injection,
// starts the Gin HTTP server, and handles graceful shutdown on SIGINT/SIGTERM.
//
// @title       DevHelp API
// @version     1.0.0
// @description A stateless developer toolbox providing HTTP inspection, cURL conversion, JWT toolkit, and JSON generation.
// @host        localhost:8080
// @BasePath    /
// @schemes     http https
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "devhelp/docs" // swagger generated docs
	"devhelp/internal/config"
	"devhelp/internal/handlers"
	"devhelp/internal/logger"
	"devhelp/internal/routes"
	"devhelp/internal/services/curlconverter"
	"devhelp/internal/services/httpinspector"
	"devhelp/internal/services/jsongenerator"
	"devhelp/internal/services/jwttool"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// ── 1. Determine environment ─────────────────────────────────────────────
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// ── 2. Load configuration ────────────────────────────────────────────────
	cfg, err := config.Load(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── 3. Initialise logger ─────────────────────────────────────────────────
	log, err := logger.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialise logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting DevHelp",
		zap.String("env", cfg.App.Env),
		zap.String("version", cfg.App.Version),
		zap.String("address", cfg.Address()),
	)

	// ── 4. Configure Gin ─────────────────────────────────────────────────────
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New() // No default middleware — we add our own.

	// ── 5. Wire services ─────────────────────────────────────────────────────
	httpInspectorSvc := httpinspector.NewService()
	curlConverterSvc := curlconverter.NewService()
	jwtSvc := jwttool.NewService()
	jsonGenSvc := jsongenerator.NewService()

	// ── 6. Wire handlers ─────────────────────────────────────────────────────
	healthH := handlers.NewHealthHandler(cfg)
	httpInspectorH := handlers.NewHTTPInspectorHandler(httpInspectorSvc)
	curlConverterH := handlers.NewCurlConverterHandler(curlConverterSvc)
	jwtH := handlers.NewJWTHandler(jwtSvc)
	jsonGenH := handlers.NewJSONGeneratorHandler(jsonGenSvc)

	// ── 7. Register routes ───────────────────────────────────────────────────
	routes.Setup(r, cfg, log, healthH, httpInspectorH, curlConverterH, jwtH, jsonGenH)

	// ── 8. Start HTTP server with graceful shutdown ───────────────────────────
	srv := &http.Server{
		Addr:         cfg.Address(),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.App.TimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.App.TimeoutSeconds)*time.Second + 5*time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server listening", zap.String("addr", cfg.Address()))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", zap.Error(err))
		os.Exit(1)
	case sig := <-quit:
		log.Info("received shutdown signal", zap.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
