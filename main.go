package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	TargetURL        string        `yaml:"target_url" env:"TARGET_URL"`
	Timeout          time.Duration `yaml:"timeout"`
	Concurrency      int           `yaml:"concurrency"`
	HTTPPort         string        `yaml:"http_port" env:"HTTP_PORT"`
	ScheduleInterval time.Duration `yaml:"schedule_interval"`
}

func main() {
	var cfg Config
	cfg.HTTPPort = ":8080"
	cfg.ScheduleInterval = 1 * time.Minute
	cfg.Timeout = 5 * time.Second
	cfg.Concurrency = 10
	cfg.TargetURL = "http://httpbin.org/ip"

	_ = cleanenv.ReadConfig("config.yaml", &cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics := NewMetricsTracker()
	fetcher := NewMultiFetcher(
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt", SOCKS5),
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt", HTTP),
		NewProxyScrapeFetcher(),
		NewFileFetcher("HTTP.txt", HTTP),
	)

	repo := NewFileRepository()
	checker := NewChecker(cfg.TargetURL, cfg.Timeout)
	engine := NewEngine(checker, repo, cfg.Concurrency, metrics)

	scheduler := NewScheduler(fetcher, engine, cfg.ScheduleInterval)

	go scheduler.RunLoop(ctx)

	server := NewServer(scheduler, metrics, ctx)
	httpServer := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: server.Routes(),
	}

	go func() {
		fmt.Printf("HTTP API server listening on %s...\n", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nReceived shutdown signal (Ctrl+C). Initiating graceful shutdown...")

	scheduler.CancelCurrent()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("HTTP server Shutdown Error: %v\n", err)
	}

	fmt.Println("Application closed cleanly.")
}