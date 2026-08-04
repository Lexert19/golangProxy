package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Fetcher interface {
	Fetch(ctx context.Context) ([]Proxy, error)
}

type FileFetcher struct {
	filePath string
	protocol Protocol
}

func NewFileFetcher(filePath string, protocol Protocol) *FileFetcher {
	return &FileFetcher{filePath: filePath, protocol: protocol}
}

func (f *FileFetcher) Fetch(ctx context.Context) ([]Proxy, error) {
	file, err := os.Open(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	var proxies []Proxy
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			var port int
			_, err := fmt.Sscanf(parts[1], "%d", &port)
			if err != nil {
				continue
			}
			proxies = append(proxies, Proxy{
				IP:       parts[0],
				Port:     port,
				Protocol: f.protocol,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return proxies, nil
}

type URLTextFetcher struct {
	url      string
	protocol Protocol
	client   *http.Client
}

func NewURLTextFetcher(url string, protocol Protocol) *URLTextFetcher {
	return &URLTextFetcher{
		url:      url,
		protocol: protocol,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *URLTextFetcher) Fetch(ctx context.Context) ([]Proxy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", f.url, err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", f.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, f.url)
	}

	var proxies []Proxy
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			var port int
			_, err := fmt.Sscanf(parts[1], "%d", &port)
			if err != nil {
				continue
			}
			proxies = append(proxies, Proxy{
				IP:       parts[0],
				Port:     port,
				Protocol: f.protocol,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %w", f.url, err)
	}

	return proxies, nil
}

type ProxyScrapeFetcher struct {
	url    string
	client *http.Client
}

type proxyScrapeResponse struct {
	Proxies []struct {
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	} `json:"proxies"`
}

func NewProxyScrapeFetcher() *ProxyScrapeFetcher {
	return &ProxyScrapeFetcher{
		url:    "https://api.proxyscrape.com/v3/free-proxy-list/get?request=displayproxies&proxy_format=protocolipport&format=json",
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *ProxyScrapeFetcher) Fetch(ctx context.Context) ([]Proxy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for ProxyScrape: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from ProxyScrape: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from ProxyScrape", resp.StatusCode)
	}

	var data proxyScrapeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode ProxyScrape JSON: %w", err)
	}

	var proxies []Proxy
	for _, item := range data.Proxies {
		proto := Protocol(strings.ToUpper(item.Protocol))
		proxies = append(proxies, Proxy{
			IP:       item.IP,
			Port:     item.Port,
			Protocol: proto,
		})
	}

	return proxies, nil
}

type MultiFetcher struct {
	fetchers []Fetcher
}

func NewMultiFetcher(fetchers ...Fetcher) *MultiFetcher {
	return &MultiFetcher{fetchers: fetchers}
}

func (m *MultiFetcher) Fetch(ctx context.Context) ([]Proxy, error) {
	var allProxies []Proxy

	for _, fetcher := range m.fetchers {
		proxies, err := fetcher.Fetch(ctx)
		if err != nil {
			fmt.Printf("Fetcher error: %v\n", err)
			continue
		}
		allProxies = append(allProxies, proxies...)
	}

	return allProxies, nil
}