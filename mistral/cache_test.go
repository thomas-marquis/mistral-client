package mistral_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/mistral-client/mistral"
	"github.com/thomas-marquis/mistral-client/mocks"
	"go.uber.org/mock/gomock"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

type cachedDataEq struct {
	t        *testing.T
	expected mistral.CachedData
}

var _ gomock.Matcher = &cachedDataEq{}

func newCachedDataEq(t *testing.T, expected mistral.CachedData) *cachedDataEq {
	return &cachedDataEq{t: t, expected: expected}
}

func (m *cachedDataEq) Matches(actual any) bool {
	if actualCachedData, ok := actual.(mistral.CachedData); ok {
		return assert.Equal(m.t, actualCachedData.Key, m.expected.Key) &&
			assert.Equal(m.t, actualCachedData.CompletionChunks, m.expected.CompletionChunks) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionRequest, m.expected.ChatCompletionRequest) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionResponse, m.expected.ChatCompletionResponse) &&
			assert.Equal(m.t, actualCachedData.EmbeddingRequest, m.expected.EmbeddingRequest) &&
			assert.Equal(m.t, actualCachedData.EmbeddingResponse, m.expected.EmbeddingResponse)
	} else if jsonData, ok := actual.([]byte); ok {
		var actualCachedData mistral.CachedData
		err := json.Unmarshal(jsonData, &actualCachedData)
		return assert.NoError(m.t, err) &&
			assert.Equal(m.t, actualCachedData.Key, m.expected.Key) &&
			assert.Equal(m.t, actualCachedData.CompletionChunks, m.expected.CompletionChunks) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionRequest, m.expected.ChatCompletionRequest) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionResponse, m.expected.ChatCompletionResponse) &&
			assert.Equal(m.t, actualCachedData.EmbeddingRequest, m.expected.EmbeddingRequest) &&
			assert.Equal(m.t, actualCachedData.EmbeddingResponse, m.expected.EmbeddingResponse)
	}

	return false
}

func (m *cachedDataEq) String() string {
	return "cached expected are equal"
}

func TestCachedDataEq(t *testing.T) {
	t.Run("should assert true with bytes input", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockEngine := mocks.NewMockEngine(ctrl)

		data := mistral.CachedData{
			CreatedAt: time.Now(),
			ChatCompletionRequest: mistral.NewChatCompletionRequest("mistral-tiny",
				[]mistral.ChatMessage{
					mistral.NewUserMessageFromString("Say hello"),
				}),
			ChatCompletionResponse: &mistral.ChatCompletionResponse{
				Choices: []mistral.ChatCompletionChoice{
					{Message: mistral.NewAssistantMessageFromString("Hello")},
				},
			},
		}

		jsonData, _ := json.Marshal(data)

		mockEngine.EXPECT().
			Set(gomock.AssignableToTypeOf(ctxType), "12345", newCachedDataEq(t, data)).
			Return(nil).
			Times(1)

		ctx := context.TODO()

		// When
		mockEngine.Set(ctx, "12345", jsonData)
	})
}

func TestCachedClientDecorator_ChatCompletion(t *testing.T) {
	t.Run("should return cached request and never call the API", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		cacheKey := "13293addba190273d98d2a572838b15c3202384f98333068afdc5e42f1ef1481"
		expectedResp := &mistral.ChatCompletionResponse{
			Choices: []mistral.ChatCompletionChoice{
				{Message: mistral.NewAssistantMessageFromString("Hello")},
			},
		}

		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		expectedCachedData := mistral.CachedData{
			ChatCompletionResponse: expectedResp,
			ChatCompletionRequest:  req,
		}
		cachedJson, _ := json.Marshal(expectedCachedData)

		mockEngine.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq(cacheKey)).
			Return(cachedJson, nil).
			Times(1)

		mockClient.EXPECT().
			ChatCompletion(gomock.Any(), gomock.Any()).
			Times(0)

		// When
		res, err := c.ChatCompletion(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, res)
	})

	t.Run("should call API when cache miss then save it", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		expectedResp := &mistral.ChatCompletionResponse{
			Choices: []mistral.ChatCompletionChoice{
				{Message: mistral.NewAssistantMessageFromString("Hello")},
			},
		}

		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		cacheKey := "13293addba190273d98d2a572838b15c3202384f98333068afdc5e42f1ef1481"
		expectedCacheData := mistral.CachedData{
			Key:                    cacheKey,
			CreatedAt:              time.Now(),
			ChatCompletionRequest:  req,
			ChatCompletionResponse: expectedResp,
		}

		mockEngine.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockClient.EXPECT().
			ChatCompletion(gomock.AssignableToTypeOf(ctxType), gomock.Eq(req)).
			Return(expectedResp, nil).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.AssignableToTypeOf(ctxType), gomock.Eq(cacheKey), newCachedDataEq(t, expectedCacheData)).
			Return(nil).
			Times(1)

		// When
		res, err := c.ChatCompletion(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, res)
	})

	t.Run("should return an error when input request is nil", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		mockClient.EXPECT().
			ChatCompletion(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		// When
		res, err := c.ChatCompletion(ctx, nil)

		// Then
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
		assert.ErrorContains(t, err, "request cannot be nil")
		assert.Nil(t, res)
	})

	t.Run("should return error on client ChatCompletion failure", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some error")

		mockClient.EXPECT().
			ChatCompletion(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		ctx := context.TODO()
		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		// When
		_, err := c.ChatCompletion(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("should return a cache failure error when engine set failed", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some cache set error")

		resp := &mistral.ChatCompletionResponse{
			Choices: []mistral.ChatCompletionChoice{
				{Message: mistral.NewAssistantMessageFromString("Hello")},
			},
		}

		mockClient.EXPECT().
			ChatCompletion(gomock.Any(), gomock.Any()).
			Return(resp, nil).
			Times(1)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr).
			Times(1)

		ctx := context.TODO()
		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		// When
		_, err := c.ChatCompletion(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
	})

	t.Run("should return a cache failure error when engine Get failed", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some cache get error")

		mockClient.EXPECT().
			ChatCompletion(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		ctx := context.TODO()
		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		// When
		_, err := c.ChatCompletion(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
	})
}

func TestCachedClientDecorator_Embeddings(t *testing.T) {
	t.Run("should return cached request and never call the API", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		cacheKey := "f85ff9609db6e0a65d1cff1b05188bca6de5e7dcb3e9f3c02a2579765fc4b063"
		expectedResp := &mistral.EmbeddingResponse{
			Data: []mistral.EmbeddingData{
				{Embedding: []float32{0.1, 0.2, 0.3}},
			},
		}

		req := mistral.NewEmbeddingRequest("mistral-embed", []string{"hello"})

		expectedCachedData := mistral.CachedData{
			Key:               cacheKey,
			CreatedAt:         time.Now(),
			EmbeddingResponse: expectedResp,
			EmbeddingRequest:  req,
		}
		cachedJson, _ := json.Marshal(expectedCachedData)

		mockEngine.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Eq(cacheKey)).
			Return(cachedJson, nil).
			Times(1)

		mockClient.EXPECT().
			Embeddings(gomock.Any(), gomock.Any()).
			Times(0)

		// When
		res, err := c.Embeddings(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, res)
	})

	t.Run("should call API when cache miss then save it", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		expectedResp := &mistral.EmbeddingResponse{
			Data: []mistral.EmbeddingData{
				{Embedding: []float32{0.1, 0.2, 0.3}},
			},
		}

		req := mistral.NewEmbeddingRequest("mistral-embed", []string{"hello"})

		cacheKey := "f85ff9609db6e0a65d1cff1b05188bca6de5e7dcb3e9f3c02a2579765fc4b063"
		expectedCacheData := mistral.CachedData{
			Key:               cacheKey,
			CreatedAt:         time.Now(),
			EmbeddingRequest:  req,
			EmbeddingResponse: expectedResp,
		}

		mockEngine.EXPECT().
			Get(gomock.AssignableToTypeOf(ctxType), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockClient.EXPECT().
			Embeddings(gomock.AssignableToTypeOf(ctxType), gomock.Eq(req)).
			Return(expectedResp, nil).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.AssignableToTypeOf(ctxType), gomock.Eq(cacheKey), newCachedDataEq(t, expectedCacheData)).
			Return(nil).
			Times(1)

		// When
		res, err := c.Embeddings(ctx, req)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, res)
	})

	t.Run("should return an error when input request is nil", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		mockClient.EXPECT().
			Embeddings(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()

		// When
		res, err := c.Embeddings(ctx, nil)

		// Then
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
		assert.ErrorContains(t, err, "request cannot be nil")
		assert.Nil(t, res)
	})

	t.Run("should return error on client ChatCompletion failure", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some error")

		mockClient.EXPECT().
			Embeddings(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		ctx := context.TODO()
		req := mistral.NewEmbeddingRequest("mistral-embed", []string{"hello"})

		// When
		_, err := c.Embeddings(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("should return a cache failure error when engine set failed", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some cache set error")

		resp := &mistral.EmbeddingResponse{
			Data: []mistral.EmbeddingData{
				{Embedding: []float32{0.1, 0.2, 0.3}},
			},
		}

		mockClient.EXPECT().
			Embeddings(gomock.Any(), gomock.Any()).
			Return(resp, nil).
			Times(1)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(expectedErr).
			Times(1)

		ctx := context.TODO()
		req := mistral.NewEmbeddingRequest("mistral-embed", []string{"hello"})

		// When
		_, err := c.Embeddings(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
	})

	t.Run("should return a cache failure error when engine Get failed", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		expectedErr := errors.New("some cache get error")

		mockClient.EXPECT().
			Embeddings(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		ctx := context.TODO()
		req := mistral.NewEmbeddingRequest("mistral-embed", []string{"hello"})

		// When
		_, err := c.Embeddings(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
		assert.ErrorIs(t, err, mistral.ErrCacheFailure)
	})
}

func TestCachedClientDecorator_ChatCompletionStream(t *testing.T) {
	t.Run("should returns cached chunks", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		chunks := []*mistral.CompletionChunk{
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("Hello ")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("world!")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString(""), FinishReason: mistral.FinishReasonStop},
				},
			},
		}

		ctx := context.TODO()
		cacheKey := "74cdb708424dcd03435d4091348a9675897999d972570b23c2f9f8f5b323093f"

		req := mistral.NewChatCompletionStreamRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		cachedData := mistral.CachedData{
			Key:              cacheKey,
			CreatedAt:        time.Now(),
			CompletionChunks: chunks,
		}
		jsonData, _ := json.Marshal(cachedData)

		mockClient.EXPECT().
			ChatCompletionStream(gomock.Any(), gomock.Any()).
			Times(0)

		mockEngine.EXPECT().
			Get(gomock.Eq(ctx), gomock.Eq(cacheKey)).
			Return(jsonData, nil).
			Times(1)

		// When
		chunkChan, err := c.ChatCompletionStream(ctx, req)

		// Then
		assert.NoError(t, err)

		var receivedChunks []*mistral.CompletionChunk
		done := make(chan struct{})

		go func() {
			for chunk := range chunkChan {
				receivedChunks = append(receivedChunks, chunk)
			}
			close(done)
		}()

		assert.Eventually(t, func() bool {
			select {
			case <-done:
				return assert.Equal(t, len(chunks), len(receivedChunks)) &&
					assert.Equal(t, chunks, receivedChunks)
			}
		}, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("should actually call API on cache miss and then save chunks with cache engine", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		chunks := []*mistral.CompletionChunk{
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("Hello ")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("world!")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString(""), FinishReason: mistral.FinishReasonStop},
				},
			},
		}

		ctx := context.TODO()
		req := mistral.NewChatCompletionStreamRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		mockClient.EXPECT().
			ChatCompletionStream(gomock.Eq(ctx), gomock.Eq(req)).
			DoAndReturn(func(ctx context.Context, req *mistral.ChatCompletionRequest) (<-chan *mistral.CompletionChunk, error) {
				res := make(chan *mistral.CompletionChunk)
				go func() {
					for _, chunk := range chunks {
						res <- chunk
					}
					close(res)
				}()
				return res, nil
			}).
			Times(1)

		cacheKey := "74cdb708424dcd03435d4091348a9675897999d972570b23c2f9f8f5b323093f"

		cachedData := mistral.CachedData{
			Key:                   cacheKey,
			CreatedAt:             time.Now(),
			CompletionChunks:      chunks,
			ChatCompletionRequest: req,
		}

		mockEngine.EXPECT().
			Get(gomock.Eq(ctx), gomock.Eq(cacheKey)).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Eq(ctx), gomock.Eq(cacheKey), newCachedDataEq(t, cachedData)).
			Return(nil).
			Times(1)

		// When
		chunkChan, err := c.ChatCompletionStream(ctx, req)

		// Then
		assert.NoError(t, err)

		var receivedChunks []*mistral.CompletionChunk
		done := make(chan struct{})

		go func() {
			for chunk := range chunkChan {
				receivedChunks = append(receivedChunks, chunk)
			}
			close(done)
		}()

		assert.Eventually(t, func() bool {
			select {
			case <-done:
				return assert.Equal(t, len(chunks), len(receivedChunks)) &&
					assert.Equal(t, chunks, receivedChunks)
			}
		}, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("should emmit one additional chunk with a cache failure error and an empty content when engine Set failed", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		chunks := []*mistral.CompletionChunk{
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("Hello ")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString("world!")},
				},
			},
			{
				Choices: []mistral.CompletionResponseStreamChoice{
					{Delta: mistral.NewAssistantMessageFromString(""), FinishReason: mistral.FinishReasonStop},
				},
			},
		}

		ctx := context.TODO()
		req := mistral.NewChatCompletionStreamRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		engSetErr := errors.New("some cache set error")

		mockClient.EXPECT().
			ChatCompletionStream(gomock.Eq(ctx), gomock.Eq(req)).
			DoAndReturn(func(ctx context.Context, req *mistral.ChatCompletionRequest) (<-chan *mistral.CompletionChunk, error) {
				res := make(chan *mistral.CompletionChunk)
				go func() {
					for _, chunk := range chunks {
						res <- chunk
					}
					close(res)
				}()
				return res, nil
			}).
			Times(1)

		cachedData := mistral.CachedData{
			Key:                   "74cdb708424dcd03435d4091348a9675897999d972570b23c2f9f8f5b323093f",
			CreatedAt:             time.Now(),
			CompletionChunks:      chunks,
			ChatCompletionRequest: req,
		}

		mockEngine.EXPECT().
			Get(gomock.Eq(ctx), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Eq(ctx), gomock.Any(), newCachedDataEq(t, cachedData)).
			Return(engSetErr).
			Times(1)

		// When
		chunkChan, err := c.ChatCompletionStream(ctx, req)

		// Then
		assert.NoError(t, err)

		var lastChunk *mistral.CompletionChunk
		done := make(chan struct{})

		go func() {
			for chunk := range chunkChan {
				lastChunk = chunk
			}
			close(done)
		}()

		assert.Eventually(t, func() bool {
			select {
			case <-done:
				return assert.Error(t, lastChunk.Error) &&
					assert.ErrorIs(t, lastChunk.Error, engSetErr) &&
					assert.ErrorIs(t, lastChunk.Error, mistral.ErrCacheFailure) &&
					assert.Len(t, lastChunk.Choices, 1) &&
					assert.Empty(t, lastChunk.Choices[0].Delta.Content().String())
			}
		}, 2*time.Second, 100*time.Millisecond)
	})

	t.Run("should return error on client ChatCompletionStream failure", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockClient := mocks.NewMockClient(ctrl)
		mockEngine := mocks.NewMockEngine(ctrl)

		c := mistral.NewCached(mockClient, mockEngine)

		ctx := context.TODO()
		req := mistral.NewChatCompletionStreamRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		expectedErr := errors.New("some error")

		mockClient.EXPECT().
			ChatCompletionStream(gomock.Eq(ctx), gomock.Eq(req)).
			Return(nil, expectedErr).
			Times(1)

		mockEngine.EXPECT().
			Get(gomock.Any(), gomock.Any()).
			Return(nil, mistral.ErrCacheMiss).
			Times(1)

		mockEngine.EXPECT().
			Set(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		// When
		_, err := c.ChatCompletionStream(ctx, req)

		// Then
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCachedClientDecorator_ListModels(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockClient(ctrl)
	mockEngine := mocks.NewMockEngine(ctrl)

	c := mistral.NewCached(mockClient, mockEngine)

	models := []*mistral.BaseModelCard{
		{Name: "mistral-small-latest"},
		{Name: "mistral-large-latest"},
	}

	ctx := context.TODO()

	mockClient.EXPECT().
		ListModels(gomock.Eq(ctx)).
		Return(models, nil).
		Times(1)

	// When
	models, err := c.ListModels(ctx)

	// Then
	assert.NoError(t, err)
	assert.Len(t, models, 2)
}

func TestCachedClientDecorator_SearchModels(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockClient(ctrl)
	mockEngine := mocks.NewMockEngine(ctrl)

	c := mistral.NewCached(mockClient, mockEngine)

	models := []*mistral.BaseModelCard{
		{Name: "mistral-small-latest"},
	}

	ctx := context.TODO()

	in := &mistral.ModelCapabilities{CompletionChat: true}

	mockClient.EXPECT().
		SearchModels(gomock.Eq(ctx), gomock.Eq(in)).
		Return(models, nil).
		Times(1)

	// When
	models, err := c.SearchModels(ctx, in)

	// Then
	assert.NoError(t, err)
	assert.Len(t, models, 1)
}

func TestCachedClientDecorator_GetModel(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockClient(ctrl)
	mockEngine := mocks.NewMockEngine(ctrl)

	c := mistral.NewCached(mockClient, mockEngine)

	ctx := context.TODO()

	model := &mistral.BaseModelCard{Name: "mistral-small-latest"}

	mockClient.EXPECT().
		GetModel(gomock.Eq(ctx), gomock.Eq("mistral-small-latest")).
		Return(model, nil).
		Times(1)

	// When
	res, err := c.GetModel(ctx, "mistral-small-latest")

	// Then
	assert.NoError(t, err)
	assert.Equal(t, model, res)
}
