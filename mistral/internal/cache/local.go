package cache

import (
	"context"
	"errors"
	"os"
)

type localFsEngine struct {
	cacheDir string
}

func NewLocalFsEngine(cacheDir string) (Engine, error) {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, errors.Join(ErrCacheFailure, errors.New("cache dir creation failed"), err)
	}
	return &localFsEngine{cacheDir: cacheDir}, nil
}

func (e *localFsEngine) Get(ctx context.Context, key string) ([]byte, error) {
	return os.ReadFile(e.cacheDir + "/" + key)
}

func (e *localFsEngine) Set(ctx context.Context, key string, data []byte) error {
	return os.WriteFile(e.cacheDir+"/"+key, data, 0644)
}
