# Caching

The Mistral client provides a built-in caching system to reduce latency and save costs by avoiding redundant API calls for identical requests. 

This feature is particularly useful during **local development**: it allows you to run your project multiple times without performing the same API call each time, saving both time and API credits.

When enabled, the client stores the responses of `ChatCompletion`, `ChatCompletionStream`, and `Embeddings` locally.

## Enabling the cache

You can enable the local file system cache when initializing the client using options.

### Use the default cache directory

By default, the cache is stored in `./mistral/cache`.

```go
client := mistral.New("YOUR_API_KEY", mistral.WithLocalCache())
```

### Specify a custom cache directory

You can also provide a custom path for the cache directory:

```go
client := mistral.New("YOUR_API_KEY", mistral.WithCacheDir("./my/custom/cache"))
```

## How it works

The caching system uses a hash of the request to identify cached responses. 
1. When a request is made, the client computes a SHA-256 hash of the request parameters.
2. It checks if a file named `<hash>.json` exists in the cache directory.
3. If it exists, the client returns the cached data without calling the Mistral API.
4. If it doesn't exist (cache miss), the client calls the API, saves the response in the cache directory, and returns it.

## Example usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thomas-marquis/mistral-client/mistral"
)

func main() {
	// Enable cache
	client := mistral.New("YOUR_API_KEY", mistral.WithLocalCache())

	ctx := context.Background()
	req := mistral.NewChatCompletionRequest("mistral-small-latest",
		[]mistral.ChatMessage{
			mistral.NewUserMessageFromString("What is the capital of France?"),
		})

	// First call - will hit the API and store the result
	start := time.Now()
	res1, _ := client.ChatCompletion(ctx, req)
	fmt.Printf("First call latency: %v\n", time.Since(start))

	// Second call - will return the cached result almost instantly
	start = time.Now()
	res2, _ := client.ChatCompletion(ctx, req)
	fmt.Printf("Second call latency: %v\n", time.Since(start))
}
```

## Cache Engines

Currently, only the `localFsEngine` is implemented, which stores data as JSON files on the local disk. 
Future versions may include support for other backends like Redis or S3.
