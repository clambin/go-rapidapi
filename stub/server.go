package stub

import (
	"net/http"
	"time"
)

// Server emulates a RapidAPI endpoint
type Server struct {
	APIKey    string
	Processor func(w http.ResponseWriter, req *http.Request)
	Called    map[string]int
}

// Handle implements the RapidAPI endpoint.  It validates that the x-rapidapi-key and then calls Handler.Processor.
// The endpoint "/slow" sleeps for 60 seconds, allowing timeout conditions to be tested.
func (server *Server) Handle(w http.ResponseWriter, req *http.Request) {
	key := req.Header.Get("x-rapidapi-key")

	if key != server.APIKey {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	server.count(req.URL.Path)

	if req.URL.Path != "/timeout" {
		server.Processor(w, req)
		return
	}
loop:
	for {
		select {
		case <-req.Context().Done():
			http.Error(w, "context exceeded", http.StatusRequestTimeout)
			break loop
		case <-time.After(60 * time.Second):
			break loop
		}
	}
}

func (server *Server) count(path string) {
	if server.Called == nil {
		server.Called = make(map[string]int)
	}
	count, _ := server.Called[path]
	count++
	server.Called[path] = count
}
