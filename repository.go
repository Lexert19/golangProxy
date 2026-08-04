package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type Repository interface {
	Save(proxy Proxy) error
}

type FileRepository struct {
	mu sync.Mutex
}

func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

func (r *FileRepository) Save(p Proxy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fileName := fmt.Sprintf("%s.txt", strings.ToUpper(string(p.Protocol)))
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()

	_, err = file.WriteString(p.Address() + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}