package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxyScrapeFetcher_Fetch_MockServer(t *testing.T) {
	mockJSON := `{
		"proxies": [
			{"ip": "192.168.1.1", "port": 8080, "protocol": "http"},
			{"ip": "10.0.0.1", "port": 1080, "protocol": "socks5"}
		]
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockJSON))
	}))
	defer mockServer.Close()

	fetcher := NewProxyScrapeFetcher()
	fetcher.url = mockServer.URL

	ctx := context.Background()
	proxies, err := fetcher.Fetch(ctx)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(proxies) != 2 {
		t.Fatalf("expected 2 proxies, got: %d", len(proxies))
	}

	if proxies[0].IP != "192.168.1.1" || proxies[0].Port != 8080 || proxies[0].Protocol != HTTP {
		t.Errorf("unexpected first proxy data: %+v", proxies[0])
	}

	if proxies[1].IP != "10.0.0.1" || proxies[1].Port != 1080 || proxies[1].Protocol != SOCKS5 {
		t.Errorf("unexpected second proxy data: %+v", proxies[1])
	}
}

func TestProxyScrapeFetcher_Fetch_LiveAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live API test in -short mode")
	}

	fetcher := NewProxyScrapeFetcher()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	proxies, err := fetcher.Fetch(ctx)

	if err != nil {
		t.Fatalf("failed to fetch from live ProxyScrape API: %v", err)
	}

	if len(proxies) == 0 {
		t.Error("expected ProxyScrape to return at least one proxy, got 0")
	}

	t.Logf("successfully fetched %d proxies from ProxyScrape", len(proxies))
}

func TestMultiFetcher_Fetch(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"proxies": [{"ip": "1.1.1.1", "port": 80, "protocol": "http"}]}`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"proxies": [{"ip": "2.2.2.2", "port": 80, "protocol": "http"}]}`))
	}))
	defer server2.Close()

	f1 := NewProxyScrapeFetcher()
	f1.url = server1.URL

	f2 := NewProxyScrapeFetcher()
	f2.url = server2.URL

	multi := NewMultiFetcher(f1, f2)

	proxies, err := multi.Fetch(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(proxies) != 2 {
		t.Errorf("expected 2 proxies, got: %d", len(proxies))
	}
}