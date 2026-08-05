package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

type Engine struct {
	checker     *Checker
	repository  Repository
	concurrency int
	metrics     *MetricsTracker
}

func NewEngine(checker *Checker, repository Repository, concurrency int, metrics *MetricsTracker) *Engine {
	return &Engine{
		checker:     checker,
		repository:  repository,
		concurrency: concurrency,
		metrics:     metrics,
	}
}

func (e *Engine) Run(ctx context.Context, proxies []Proxy) error {
	e.metrics.StartRun(len(proxies))
	defer e.metrics.FinishRun()

	jobs := make(chan Proxy, len(proxies))
	for _, p := range proxies {
		jobs <- p
	}
	close(jobs)

	g, gCtx := errgroup.WithContext(ctx)

	fmt.Printf("Starting proxy checker [%d workers]...\n", e.concurrency)

	for i := 0; i < e.concurrency; i++ {
		g.Go(func() error {
			e.metrics.IncWorker()
			defer e.metrics.DecWorker()

			for {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				case p, ok := <-jobs:
					if !ok {
						return nil
					}

					start := time.Now()
					err := e.checker.Check(gCtx, p)
					dur := time.Since(start)

					e.metrics.RecordCheck(p, err, dur)

					if err != nil {
						continue
					}

					fmt.Printf("[OK] Proxy %s (%s) is working!\n", p.Address(), p.Protocol)
					if saveErr := e.repository.Save(p); saveErr != nil {
						fmt.Printf("Failed to save proxy: %v\n", saveErr)
					}
				}
			}
		})
	}

	err := g.Wait()
	fmt.Println("Finished verifying proxy list.")
	return err
}