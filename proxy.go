package main

import (
	"errors"
	"fmt"
)

type Protocol string

const (
	HTTP   Protocol = "HTTP"
	HTTPS  Protocol = "HTTPS"
	SOCKS4 Protocol = "SOCKS4"
	SOCKS5 Protocol = "SOCKS5"
)

type Proxy struct {
	IP       string
	Port     int
	Protocol Protocol
}

func (p Proxy) Address() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

var (
	ErrProxyTimeout        = errors.New("connection timeout")
	ErrBadStatusCode       = errors.New("target server returned non-200 status code")
	ErrUnsupportedProtocol = errors.New("unsupported proxy protocol")
	ErrConnectionRefused   = errors.New("connection refused by proxy")
)