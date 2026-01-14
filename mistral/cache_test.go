package mistral_test

import (
	"context"
	"encoding/json"
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
		return assert.EqualValues(m.t, actualCachedData.ChatCompletionRequest, m.expected.ChatCompletionRequest) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionResponse, m.expected.ChatCompletionResponse)
	} else if jsonData, ok := actual.([]byte); ok {
		var actualCachedData mistral.CachedData
		err := json.Unmarshal(jsonData, &actualCachedData)
		return assert.NoError(m.t, err) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionRequest, m.expected.ChatCompletionRequest) &&
			assert.Equal(m.t, actualCachedData.ChatCompletionResponse, m.expected.ChatCompletionResponse)
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

		c := mistral.Cached(mockClient, mockEngine)

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

		c := mistral.Cached(mockClient, mockEngine)

		ctx := context.TODO()

		expectedResp := &mistral.ChatCompletionResponse{
			Choices: []mistral.ChatCompletionChoice{
				{Message: mistral.NewAssistantMessageFromString("Hello")},
			},
		}

		req := mistral.NewChatCompletionRequest("mistral-tiny",
			[]mistral.ChatMessage{mistral.NewUserMessageFromString("Say hello")})

		expectedCacheData := mistral.CachedData{
			ChatCompletionRequest:  req,
			ChatCompletionResponse: expectedResp,
		}
		cacheKey := "13293addba190273d98d2a572838b15c3202384f98333068afdc5e42f1ef1481"

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
}

//func TestComputeHashFromCompletionRequest(t *testing.T) {
//	req1 := &ChatCompletionRequest{
//		Model: "mistral-tiny",
//		Messages: []ChatMessage{
//			NewUserMessageFromString("Hello"),
//		},
//	}
//
//	req2 := &ChatCompletionRequest{
//		Model: "mistral-tiny",
//		Messages: []ChatMessage{
//			NewUserMessageFromString("Hello"),
//		},
//	}
//
//	req3 := &ChatCompletionRequest{
//		Model: "mistral-small",
//		Messages: []ChatMessage{
//			NewUserMessageFromString("Hello"),
//		},
//	}
//
//	hash1, err := computeHashFromCompletionRequest(req1)
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	hash2, err := computeHashFromCompletionRequest(req2)
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	hash3, err := computeHashFromCompletionRequest(req3)
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	if hash1 == "" {
//		t.Error("hash1 should not be empty")
//	}
//
//	if hash1 != hash2 {
//		t.Errorf("hash1 and hash2 should be equal, got %s and %s", hash1, hash2)
//	}
//
//	if hash1 == hash3 {
//		t.Errorf("hash1 and hash3 should be different, got %s", hash1)
//	}
//}
//
//func TestComputeHashFromCompletionRequest_Nil(t *testing.T) {
//	_, err := computeHashFromCompletionRequest(nil)
//	if err == nil {
//		t.Error("expected error for nil request, got nil")
//	}
//}
