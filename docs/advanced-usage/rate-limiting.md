# Rate Limiting

An option is available to limit the number of requests per second.

You have to specify it once on the client initialization. 
Then this limiter will be applied to all requests (chat completion, embeddings...).
Except for listing or searching for models.

The limiter is based on [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate).

```go
import (
	...
    "time"
    
    "golang.org/x/time/rate"
)

...

rl := rate.NewLimiter(rate.Every(10*time.Second), 50) // 50 request every 10 second
client := mistral.New(apiKey, mistral.WithRateLimiter(rl))
```

Learn more about rate limiting with `golang.org/x/time/rate` [in this cool article](https://medium.com/mflow/rate-limiting-in-golang-http-client-a22fba15861a).