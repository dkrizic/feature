# HTTP Client with OpenTelemetry Instrumentation

This package provides an instrumented HTTP client that automatically traces outgoing HTTP requests using OpenTelemetry.

## Usage

### Using the Default Client

The simplest way to use the instrumented HTTP client is through the package-level functions:

```go
import (
    "context"
    "github.com/prodyna-yasm/yasm-external/telemetry/httpclient"
)

// GET request
resp, err := httpclient.Get(ctx, "https://api.example.com/data")
if err != nil {
    // handle error
}
defer resp.Body.Close()

// POST request
body := strings.NewReader(`{"key": "value"}`)
resp, err = httpclient.Post(ctx, "https://api.example.com/data", "application/json", body)
```

### Using a Custom Client

For more control, you can create a custom client instance:

```go
// Create a new instrumented client with default transport
client := httpclient.New(nil)

// Or with a custom transport
customTransport := &http.Transport{
    MaxIdleConns: 100,
    // ... other settings
}
client := httpclient.New(customTransport)

// Use the client
resp, err := client.Get(ctx, "https://api.example.com/data")
```

### Making Custom Requests

For more complex requests, use the `Do` method:

```go
client := httpclient.DefaultClient()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/data", nil)
if err != nil {
    // handle error
}

// Add custom headers
req.Header.Set("Authorization", "Bearer token")

resp, err := client.Do(req)
```

## Features

- Automatic OpenTelemetry tracing for all outgoing HTTP requests
- Trace context propagation to downstream services
- Consistent with otelhttp instrumentation standards
- Drop-in replacement for standard http.Client methods
