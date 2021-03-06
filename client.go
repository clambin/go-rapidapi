package rapidapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// API interface for a RapiAPI client
//go:generate mockery --name API
type API interface {
	Call(endpoint string) (body []byte, err error)
	CallWithContext(ctx context.Context, endpoint string) (body []byte, err error)
}

// Client represents a RapidAPI client
//
// For unit tests, set the URL field. This will ignore the hostname field.
type Client struct {
	URL        string
	hostname   string
	apiKey     string
	httpClient *http.Client
}

// New creates a new client
func New(hostname, apiKey string) *Client {
	return &Client{
		httpClient: http.DefaultClient,
		hostname:   hostname,
		apiKey:     apiKey,
	}
}

// WithHTTPClient sets the client's httpClient
func (client *Client) WithHTTPClient(httpClient *http.Client) *Client {
	client.httpClient = httpClient
	return client
}

// Call an endpoint on the API
func (client *Client) Call(endpoint string) (body []byte, err error) {
	return client.CallWithContext(context.Background(), endpoint)
}

func (client Client) makeURL(endpoint string) string {
	baseURL := "https://" + client.hostname
	if client.URL != "" {
		baseURL = client.URL
	}
	return baseURL + endpoint
}

// CallWithContext calls an endpoint on the API with a provided context
func (client *Client) CallWithContext(ctx context.Context, endpoint string) (body []byte, err error) {
	url := client.makeURL(endpoint)

	const (
		initWaitTime = 100 * time.Millisecond
		maxWaitTime  = 5 * time.Second
		maxRetry     = 10
	)
	waitTime := initWaitTime

	for retries := 0; retries < maxRetry; retries++ {
		body, err = client.call(ctx, url)

		if err == nil {
			break
		}

		if err.Error() != fmt.Sprintf("%3d %s", http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests)) {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitTime):
			break
		}

		if waitTime < maxWaitTime {
			waitTime *= 2
		}
	}

	return
}

func (client *Client) call(ctx context.Context, url string) (body []byte, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Add("x-rapidapi-key", client.apiKey)
	req.Header.Add("x-rapidapi-host", client.hostname)

	var resp *http.Response
	resp, err = client.httpClient.Do(req)

	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return io.ReadAll(resp.Body)
}
