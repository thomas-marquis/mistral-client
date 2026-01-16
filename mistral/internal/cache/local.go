package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type localFsEngine struct {
	cacheDir string
}

func NewLocalFsEngine(cacheDir string) (Engine, error) {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("cache dir creation failed: %w", err)
	}
	return &localFsEngine{cacheDir: cacheDir}, nil
}

func (e *localFsEngine) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(e.cacheDir, key+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}
	return data, nil
}

func (e *localFsEngine) Set(ctx context.Context, key string, data []byte) error {
	err := os.WriteFile(filepath.Join(e.cacheDir, key+".json"), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}
