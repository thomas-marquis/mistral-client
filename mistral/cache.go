package mistral

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral/internal/cache"
)

const (
	DefaultCacheDir = "./.mistral/cache"
)

var (
	ErrCacheMiss    = cache.ErrCacheMiss
	ErrCacheFailure = errors.New("cache failure")
)

type cacheConfig struct {
	enabled  bool
	cacheDir string
}

type CachedData struct {
	Key                    string
	CreatedAt              time.Time
	ChatCompletionRequest  *ChatCompletionRequest
	ChatCompletionResponse *ChatCompletionResponse
	EmbeddingRequest       *EmbeddingRequest
	EmbeddingResponse      *EmbeddingResponse
	CompletionChunks       []*CompletionChunk
}

type CacheEngine interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

type cachedClientDecorator struct {
	client Client
	engine CacheEngine
}

func newCachedClient(client Client, engine CacheEngine) (Client, error) {
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
			cacheData, err := json.Marshal(CachedData{
				Key:                    cacheKey,
				CreatedAt:              time.Now(),
				ChatCompletionRequest:  request,
				ChatCompletionResponse: res,
			})
			if err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			if err := c.engine.Set(ctx, cacheKey, cacheData); err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			return res, nil
		}
		return nil, errors.Join(ErrCacheFailure, err)
	}

	var cachedData CachedData
	if err := json.Unmarshal(data, &cachedData); err != nil {
		return nil, errors.Join(ErrCacheFailure, err)
	}

	return cachedData.ChatCompletionResponse, nil
}

func (c *cachedClientDecorator) ChatCompletionStream(ctx context.Context, request *ChatCompletionRequest) (<-chan *CompletionChunk, error) {
	cacheKey, err := computeHashKey(request)
	if err != nil {
		return nil, err
	}

	data, err := c.engine.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			res, err := c.client.ChatCompletionStream(ctx, request)
			if err != nil {
				return nil, err
			}

			proxyChan := make(chan *CompletionChunk)

			go func() {
				defer close(proxyChan)
				cachedData := CachedData{
					Key:                   cacheKey,
					CreatedAt:             time.Now(),
					ChatCompletionRequest: request,
					CompletionChunks:      make([]*CompletionChunk, 0),
				}
				for chunk := range res {
					cachedData.CompletionChunks = append(cachedData.CompletionChunks, chunk)
					proxyChan <- chunk
				}
				cacheData, err := json.Marshal(cachedData)
				if err != nil {
					proxyChan <- &CompletionChunk{
						Error: errors.Join(ErrCacheFailure, err),
						Choices: []CompletionResponseStreamChoice{
							{Delta: NewAssistantMessageFromString("")},
						},
					}
				}
				if err := c.engine.Set(ctx, cacheKey, cacheData); err != nil {
					proxyChan <- &CompletionChunk{
						Error: errors.Join(ErrCacheFailure, err),
						Choices: []CompletionResponseStreamChoice{
							{Delta: NewAssistantMessageFromString("")},
						},
					}
				}
			}()
			return proxyChan, nil
		}
		return nil, errors.Join(ErrCacheFailure, err)
	}

	var cachedData CachedData
	if err := json.Unmarshal(data, &cachedData); err != nil {
		return nil, errors.Join(ErrCacheFailure, err)
	}

	resChan := make(chan *CompletionChunk)

	go func() {
		defer close(resChan)
		for _, chunk := range cachedData.CompletionChunks {
			resChan <- chunk
		}
	}()

	return resChan, nil
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
			cacheData, err := json.Marshal(CachedData{
				Key:               cacheKey,
				CreatedAt:         time.Now(),
				EmbeddingRequest:  request,
				EmbeddingResponse: res,
			})
			if err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			if err := c.engine.Set(ctx, cacheKey, cacheData); err != nil {
				return nil, errors.Join(ErrCacheFailure, err)
			}
			return res, nil
		}
		return nil, errors.Join(ErrCacheFailure, err)
	}

	var cachedData CachedData
	if err := json.Unmarshal(data, &cachedData); err != nil {
		return nil, errors.Join(ErrCacheFailure, err)
	}

	return cachedData.EmbeddingResponse, nil
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
	if in == nil || (reflect.ValueOf(in).Kind() == reflect.Ptr && reflect.ValueOf(in).IsNil()) {
		return "", errors.Join(ErrCacheFailure, errors.New("request cannot be nil"))
	}

	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
