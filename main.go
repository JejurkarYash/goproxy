package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// each server configuration
type Backend struct {
	URL          string
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock() // writing lock (so only one thread can write at time)
	defer b.mu.Unlock()
	b.Alive = alive

}

func (b *Backend) GetAlive() bool {
	b.mu.RLock() // -> reader lock (multiple thread's can read at a time )
	defer b.mu.RUnlock()
	return b.Alive
}

type ServerPool struct {
	mu       sync.Mutex
	backends []*Backend // url array
	current  int        // current url index
}

func (sp *ServerPool) GetNextBackend() *Backend {
	sp.mu.Lock() // -> why not here read lock ? -> what if 100 req hit at the same milisecond
	defer sp.mu.Unlock()
	for i := 0; i < len(sp.backends); i++ {
		idx := (sp.current + i) % len(sp.backends)
		if sp.backends[idx].GetAlive() {
			sp.current = (idx + 1) % len(sp.backends)
			return sp.backends[idx]
		}
	}
	return nil
}

// Constructor function
func NewBackend(rawURL string) (*Backend, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Backend{
		URL:          rawURL,
		Alive:        true,
		ReverseProxy: httputil.NewSingleHostReverseProxy(parsedURL),
	}, nil
}

// entry point
func main() {
	serverURL := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	var backends []*Backend

	for _, rawURL := range serverURL {
		b, err := NewBackend(rawURL)
		if err != nil {
			log.Fatal(err)
		}

		backends = append(backends, b)
	}

	// creating a server pool with bunch of servers
	NewServerPool := &ServerPool{
		backends: backends,
	}

	// starting mock backend
	go StartMockBackend(":8081")
	go StartMockBackend(":8082")
	go StartMockBackend(":8083")

	// starting a new goroutine for backend health check
	go BackendHealthCheck(NewServerPool)

	// starting Proxy server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// forwarding the request to mock backend
		backend := NewServerPool.GetNextBackend()
		backend.ReverseProxy.ServeHTTP(w, r)
	})

	proxyServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Proxy Server is listening on port:8080")
	proxyServer.ListenAndServe()

}

// starting a dummy server
func StartMockBackend(port string) {
	// router
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// seding response
		message := "Hello from Backend:" + port
		w.Write([]byte(message))
	})

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	fmt.Println("Mock Backend is running on port :", port)
	server.ListenAndServe()

}

func BackendHealthCheck(sp *ServerPool) {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		for _, b := range sp.backends {
			status := IsAliveBackend(b.URL)
			b.SetAlive(status)
		}
	}
}

func IsAliveBackend(rawURL string) bool {
	resp, err := http.Get(rawURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
