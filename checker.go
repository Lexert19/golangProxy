package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Checker struct {
	targetURL string
	timeout   time.Duration
}

func NewChecker(targetURL string, timeout time.Duration) *Checker {
	return &Checker{
		targetURL: targetURL,
		timeout:   timeout,
	}
}

func (c *Checker) Check(ctx context.Context, p Proxy) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	transport, err := c.createTransport(p)
	if err != nil {
		return fmt.Errorf("transport configuration error: %w", err)
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, c.targetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)

	if err != nil {
		if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
			return ErrProxyTimeout
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return ErrProxyTimeout
		}
		if strings.Contains(err.Error(), "connection refused") {
			return ErrConnectionRefused
		}
		return fmt.Errorf("proxy connection error: %w", err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: received status code %d", ErrBadStatusCode, resp.StatusCode)
	}

	return nil
}

func (c *Checker) createTransport(p Proxy) (*http.Transport, error) {
	switch p.Protocol {
	case HTTP, HTTPS:
		proxyURL, err := url.Parse(fmt.Sprintf("http://%s", p.Address()))
		if err != nil {
			return nil, err
		}
		return &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}, nil
	case SOCKS5:
		dialer, err := proxy.SOCKS5("tcp", p.Address(), nil, proxy.Direct)
		if err != nil {
			return nil, err
		}
		return &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}, nil

	default:
		return nil, ErrUnsupportedProtocol
	}
}