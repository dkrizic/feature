package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Client is an instrumented HTTP client that wraps http.Client with OpenTelemetry tracing.
// It automatically sets the http.route attribute when possible for outgoing requests.
type Client struct {
	client *http.Client
}

// New creates a new instrumented HTTP client with OpenTelemetry tracing.
// The base parameter can be nil, in which case http.DefaultTransport is used.
func New(base http.RoundTripper) *Client {
	if base == nil {
		base = http.DefaultTransport
	}
	transport := otelhttp.NewTransport(base)
	return &Client{
		client: &http.Client{
			Transport: transport,
		},
	}
}

// DefaultClient returns a shared instrumented HTTP client.
// This is suitable for most use cases where a custom client is not required.
func DefaultClient() *Client {
	return &Client{
		client: otelhttp.DefaultClient,
	}
}

// Do sends an HTTP request and returns an HTTP response, using the instrumented client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Get issues a GET request to the specified URL with OpenTelemetry tracing.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post issues a POST request to the specified URL with OpenTelemetry tracing.
func (c *Client) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// PostForm issues a POST request with form data to the specified URL with OpenTelemetry tracing.
func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return c.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// Head issues a HEAD request to the specified URL with OpenTelemetry tracing.
func (c *Client) Head(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get issues a GET request using the default instrumented client.
func Get(ctx context.Context, url string) (*http.Response, error) {
	return DefaultClient().Get(ctx, url)
}

// Post issues a POST request using the default instrumented client.
func Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	return DefaultClient().Post(ctx, url, contentType, body)
}

// PostForm issues a POST form request using the default instrumented client.
func PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return DefaultClient().PostForm(ctx, url, data)
}

// Head issues a HEAD request using the default instrumented client.
func Head(ctx context.Context, url string) (*http.Response, error) {
	return DefaultClient().Head(ctx, url)
}
