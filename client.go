package rapidapi

import (
	"context"
	"errors"
	"io/ioutil"
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
// APIKey should contain the RapidAPI API Key
// If Hostname is set, it will be used to construct the URL and fill in the x-rapidapi-host header field.
// For unit tests, set the URL field and ignore the Hostname field.
type Client struct {
	HTTPClient *http.Client
	Hostname   string
	URL        string
	APIKey     string
}

// Call an endpoint on the API
func (client *Client) Call(endpoint string) (body []byte, err error) {
	return client.CallWithContext(context.Background(), endpoint)
}

// CallWithContext calls an endpoint on the API with a provided context
func (client *Client) CallWithContext(ctx context.Context, endpoint string) (body []byte, err error) {
	httpClient := http.DefaultClient
	if client.HTTPClient != nil {
		httpClient = client.HTTPClient
	}

	baseURL := "https://" + client.Hostname
	if client.URL != "" {
		baseURL = client.URL
	}
	url := baseURL + endpoint

	const (
		initWaitTime = 100 * time.Millisecond
		maxWaitTime  = 5 * time.Second
		maxRetry     = 10
	)
	waitTime := initWaitTime
	retries := 0
	for calling := true; calling && retries < maxRetry; retries++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		req.Header.Add("x-rapidapi-key", client.APIKey)
		req.Header.Add("x-rapidapi-host", client.Hostname)

		var resp *http.Response
		resp, err = httpClient.Do(req)

		if err != nil {
			break
		}

		switch resp.StatusCode {
		case http.StatusOK:
			body, err = ioutil.ReadAll(resp.Body)
			calling = false
		case http.StatusTooManyRequests:
			err = errors.New(resp.Status)
			time.Sleep(waitTime)
			if waitTime < maxWaitTime {
				waitTime *= 2
			}
		default:
			err = errors.New(resp.Status)
			calling = false
		}

		_ = resp.Body.Close()
	}

	return
}
