package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	TargetURL   string        `yaml:"target_url" env:"TARGET_URL"`
	Timeout     time.Duration `yaml:"timeout"`
	Concurrency int           `yaml:"concurrency"`
}


func main() {
	var cfg Config
	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
	fetcher := NewMultiFetcher(
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt", SOCKS5),
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt", HTTP),
		NewProxyScrapeFetcher(),
		NewFileFetcher("HTTP.txt", HTTP),
	)

	repo := NewFileRepository()
	checker := NewChecker(cfg.TargetURL, cfg.Timeout)
	engine := NewEngine(checker, repo, cfg.Concurrency)


	ctx := context.Background()
	proxies, err := fetcher.Fetch(ctx)
	if err != nil {
		fmt.Printf("Critical error fetching proxies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d proxies in total.\n", len(proxies))

	engine.Run(ctx, proxies)
}