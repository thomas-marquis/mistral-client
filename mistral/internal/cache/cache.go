package cache

import (
	"context"
	"errors"
)

var (
	ErrCacheMiss    = errors.New("cache miss")
	ErrCacheFailure = errors.New("something went wrong with caching")
)

type Engine interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}
