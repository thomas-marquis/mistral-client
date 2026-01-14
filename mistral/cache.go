package mistral

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral/internal/cache"
)

const (
	DefaultCacheDir = "./mistral/cache"
)

var (
	ErrCacheMiss    = cache.ErrCacheMiss
	ErrCacheFailure = cache.ErrCacheFailure
)

type cacheConfig struct {
	enabled  bool
	cacheDir string
}

type CachedData struct {
	CreatedAt time.Time
	Request   *ChatCompletionRequest
	Response  *ChatCompletionResponse
}

type cachedClientDecorator struct {
	client Client
	engine cache.Engine
}

func newCachedClient(client Client, engine cache.Engine) (Client, error) {
	return &cachedClientDecorator{client: client, engine: engine}, nil
}

func (c *cachedClientDecorator) ChatCompletion(ctx context.Context, request *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	cacheKey, err := computeHashKey(request)
	if err != nil {
		return nil, err
	}

	data, err := c.engine.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			res, err := c.client.ChatCompletion(ctx, request)
			if err != nil {
				return nil, err
			}
			cacheData, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			if err := c.engine.Set(ctx, cacheKey, cacheData); err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	var resp ChatCompletionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.Join(ErrCacheFailure, err)
	}

	return &resp, nil
}

func (c *cachedClientDecorator) ChatCompletionStream(ctx context.Context, request *ChatCompletionRequest) (<-chan *CompletionChunk, error) {
	return nil, errors.New("not implemented yet")
}

func (c *cachedClientDecorator) Embeddings(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error) {
	cacheKey, err := computeHashKey(request)
	if err != nil {
		return nil, err
	}

	data, err := c.engine.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			res, err := c.client.Embeddings(ctx, request)
			if err != nil {
				return nil, err
			}
			cacheData, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			if err := c.engine.Set(ctx, cacheKey, cacheData); err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	var resp EmbeddingResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errors.Join(ErrCacheFailure, err)
	}

	return &resp, nil
}

func (c *cachedClientDecorator) ListModels(ctx context.Context) ([]*BaseModelCard, error) {
	return c.client.ListModels(ctx)
}

func (c *cachedClientDecorator) SearchModels(ctx context.Context, capabilities *ModelCapabilities) ([]*BaseModelCard, error) {
	return c.client.SearchModels(ctx, capabilities)
}

func (c *cachedClientDecorator) GetModel(ctx context.Context, modelId string) (*BaseModelCard, error) {
	return c.client.GetModel(ctx, modelId)
}

func computeHashKey(in any) (string, error) {
	if in == nil {
		return "", errors.New("request cannot be nil")
	}

	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
