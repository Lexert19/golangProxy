package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)


func BenchmarkParseProxyLines(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(fmt.Sprintf("192.168.1.%d:8080\n", i%255))
	}
	input := sb.String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		scanner := bufio.NewScanner(strings.NewReader(input))
		var proxies []Proxy
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				var port int
				if _, err := fmt.Sscanf(parts[1], "%d", &port); err == nil {
					proxies = append(proxies, Proxy{
						IP:       parts[0],
						Port:     port,
						Protocol: HTTP,
					})
				}
			}
		}
	}
}