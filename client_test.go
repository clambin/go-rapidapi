package rapidapi_test

import (
	"github.com/clambin/go-rapidapi"
	"github.com/clambin/go-rapidapi/stub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Call(t *testing.T) {
	server := stub.Server{
		APIKey:    "1234",
		Processor: Processor,
	}
	testServer := httptest.NewServer(http.HandlerFunc(server.Handle))
	defer testServer.Close()

	var testCases = []struct {
		name     string
		pass     bool
		apikey   string
		endpoint string
		response string
	}{
		{"happy", true, "1234", "/", "OK"},
		{"no apikey", false, "", "/", "403 Forbidden"},
		{"bad endpoint", false, "1234", "/invalid", "404 Not Found"},
	}

	client := rapidapi.Client{
		HTTPClient: &http.Client{},
		URL:        testServer.URL,
	}

	for _, testCase := range testCases {
		client.APIKey = testCase.apikey
		response, err := client.Call(testCase.endpoint)

		if testCase.pass == true {
			if assert.NoError(t, err, testCase.name) {
				assert.Equal(t, testCase.response, string(response), testCase.name)
			}
		} else {
			if assert.Error(t, err, testCase.name) {
				assert.Equal(t, testCase.response, err.Error(), testCase.name)
			}
		}

	}
}

func TestClient_Call_TimeOut(t *testing.T) {
	server := stub.Server{
		APIKey:    "1234",
		Processor: Processor,
	}
	testServer := httptest.NewServer(http.HandlerFunc(server.Handle))
	defer testServer.Close()

	client := rapidapi.Client{
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
		URL:        testServer.URL,
		APIKey:     "1234",
	}

	_, err := client.Call("/timeout")
	assert.Error(t, err)
}

func TestClient_Call_TooManyRequests(t *testing.T) {
	server := stub.Server{
		APIKey:    "1234",
		Processor: Processor,
	}
	testServer := httptest.NewServer(http.HandlerFunc(server.Handle))
	defer testServer.Close()

	client := rapidapi.Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		URL:        testServer.URL,
		APIKey:     "1234",
	}

	_, err := client.Call("/retry")
	require.NoError(t, err)
	assert.Greater(t, server.Called["/retry"], 1)
}

var first = true

func Processor(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		_, _ = w.Write([]byte("OK"))
	case "/retry":
		if first {
			first = false
			http.Error(w, "slow down!", http.StatusTooManyRequests)
		} else {
			_, _ = w.Write([]byte("OK"))
		}
	default:
		http.Error(w, "Page not found", http.StatusNotFound)
	}
}
