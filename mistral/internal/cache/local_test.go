package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalFsEngine(t *testing.T) {
	cacheDir, err := os.MkdirTemp("", "mistral-cache-test")
	assert.NoError(t, err)
	defer os.RemoveAll(cacheDir) //nolint:errcheck

	engine, err := NewLocalFsEngine(cacheDir)
	assert.NoError(t, err)

	ctx := context.Background()
	key := "test-key"
	data := []byte("test-data")

	t.Run("Set and Get", func(t *testing.T) {
		err := engine.Set(ctx, key, data)
		assert.NoError(t, err)

		// Check if file exists with .json extension
		expectedFile := filepath.Join(cacheDir, key+".json")
		assert.FileExists(t, expectedFile)

		received, err := engine.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, data, received)
	})

	t.Run("Get Cache Miss", func(t *testing.T) {
		received, err := engine.Get(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrCacheMiss)
		assert.Nil(t, received)
	})

	t.Run("NewLocalFsEngine creates dir", func(t *testing.T) {
		newDir := filepath.Join(cacheDir, "nested", "cache")
		_, err := NewLocalFsEngine(newDir)
		assert.NoError(t, err)
		assert.DirExists(t, newDir)
	})
}
