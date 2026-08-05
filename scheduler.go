package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Scheduler struct {
	fetcher  Fetcher
	engine   *Engine
	interval time.Duration
	mu       sync.Mutex
	cancel   context.CancelFunc
}

func NewScheduler(fetcher Fetcher, engine *Engine, interval time.Duration) *Scheduler {
	return &Scheduler{
		fetcher:  fetcher,
		engine:   engine,
		interval: interval,
	}
}

func (s *Scheduler) StartSingleRun(parentCtx context.Context) bool {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return false
	}

	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.cancel = nil
			s.mu.Unlock()
		}()

		s.executeRun(ctx)
	}()

	return true
}

func (s *Scheduler) CancelCurrent() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
		return true
	}
	return false
}

func (s *Scheduler) RunLoop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()


	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler loop stopped.")
			return
		case <-ticker.C:
			fmt.Println("Scheduled tick triggered.")
			if !s.StartSingleRun(ctx) {
				fmt.Println("Previous check still running, skipping scheduled run.")
			}
		}
	}
}

func (s *Scheduler) executeRun(ctx context.Context) {
	proxies, err := s.fetcher.Fetch(ctx)
	if err != nil {
		fmt.Printf("Scheduler error fetching proxies: %v\n", err)
		return
	}

	if len(proxies) == 0 {
		fmt.Println("No proxies fetched.")
		return
	}

	fmt.Printf("Fetched %d proxies. Starting execution...\n", len(proxies))
	if err := s.engine.Run(ctx, proxies); err != nil {
		fmt.Printf("Run ended with result/err: %v\n", err)
	}
}