package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	fetcher := NewMultiFetcher(
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt", SOCKS5),
		NewURLTextFetcher("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt", HTTP),
		NewProxyScrapeFetcher(),
		NewFileFetcher("HTTP.txt", HTTP),
	)

	repo := NewFileRepository()
	checker := NewChecker("http://dominik-chyziak.pl", 5*time.Second)
	engine := NewEngine(checker, repo, 200)

	ctx := context.Background()
	proxies, err := fetcher.Fetch(ctx)
	if err != nil {
		fmt.Printf("Critical error fetching proxies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d proxies in total.\n", len(proxies))

	engine.Run(ctx, proxies)
}