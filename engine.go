package main

import (
	"context"
	"fmt"
	"sync"
)

type Engine struct {
	checker     *Checker
	repository  Repository
	concurrency int
}

func NewEngine(checker *Checker, repository Repository, concurrency int) *Engine {
	return &Engine{
		checker:     checker,
		repository:  repository,
		concurrency: concurrency,
	}
}

func (e *Engine) Run(ctx context.Context, proxies []Proxy) {
	jobs := make(chan Proxy, len(proxies))
	var wg sync.WaitGroup

	fmt.Printf("Starting proxy checker [%d workers]...\n", e.concurrency)

	for i := 0; i < e.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for p := range jobs {
				err := e.checker.Check(ctx, p)
				if err != nil {
					continue
				}

				fmt.Printf("[OK] Proxy %s (%s) is working!\n", p.Address(), p.Protocol)
				if saveErr := e.repository.Save(p); saveErr != nil {
					fmt.Printf("Failed to save proxy: %v\n", saveErr)
				}
			}
		}(i)
	}

	for _, p := range proxies {
		jobs <- p
	}
	close(jobs)

	wg.Wait()
	fmt.Println("Finished verifying proxy list.")
}