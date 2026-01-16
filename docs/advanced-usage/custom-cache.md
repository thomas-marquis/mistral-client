# Custom Cache Engine

The Mistral client allows you to provide your own cache engine implementation. This is useful if you want to store cached responses in a shared storage like Redis, S3, or a database instead of the local file system.

## The CacheEngine Interface

To implement a custom cache engine, you need to satisfy the `CacheEngine` interface defined in the `mistral` package:

```go
type CacheEngine interface {
    // Get retrieves the cached data for the given key.
    // It should return mistral.ErrCacheMiss if the key is not found.
    Get(ctx context.Context, key string) ([]byte, error)
    
    // Set stores the data for the given key.
    Set(ctx context.Context, key string, data []byte) error
}
```

## Implementation Example (Conceptual)

Here is a conceptual example of how you could implement a cache engine using S3:

```go
import (
    "context"
    "github.com/thomas-marquis/mistral-client/mistral"
    // ... S3 SDK imports
)

type S3CacheEngine struct {
    bucket string
    // s3Client *s3.Client
}

func (e *S3CacheEngine) Get(ctx context.Context, key string) ([]byte, error) {
    // 1. Check if the object exists in S3 (e.g., using HeadObject)
    // 2. If it doesn't exist, return mistral.ErrCacheMiss
    // 3. If it exists, download and return the bytes
    return nil, mistral.ErrCacheMiss // Placeholder
}

func (e *S3CacheEngine) Set(ctx context.Context, key string, data []byte) error {
    // 1. Upload the data to S3 using the key
    return nil
}
```

## Using your Custom Engine

Once your engine is implemented, wrap your Mistral client using the `NewCached` function:

```go
func main() {
    // Initialize your standard client
    client := mistral.New("YOUR_API_KEY")

    // Initialize your custom engine
    myEngine := &S3CacheEngine{bucket: "my-mistral-cache"}

    // Wrap the client with your custom engine
    cachedClient := mistral.NewCached(client, myEngine)

    // All calls through cachedClient will now use your custom S3 cache
    req := mistral.NewChatCompletionRequest("mistral-small-latest", ...)
    res, err := cachedClient.ChatCompletion(context.Background(), req)
}
```

## Error Handling

When implementing `Get`, it is important to return `mistral.ErrCacheMiss` when the requested key is not found in your cache. This tells the client to proceed with the actual API call and then store the result using the `Set` method.
