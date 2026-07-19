package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// entry point
func main() {

	// loading config from config.json file
	Config := LoadConfig()

	var backends []*Backend

	for _, rawURL := range Config.Backends {
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

	// attaching the error handler
	for _, b := range NewServerPool.backends {

		b.ReverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {

			// logging to terminal
			log.Printf("[%s] connection failed: %v", b.URL, err)
			//  marking it as unhealthy
			b.SetAlive(false)

			nextBackend := NewServerPool.GetNextBackend()
			if nextBackend != nil {
				log.Printf("[%s] connection retrying...\n", nextBackend.URL)
				nextBackend.ReverseProxy.ServeHTTP(w, r)

				return
			}

			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
	}

	// starting mock backend
	go StartMockBackend(":8081")
	go StartMockBackend(":8082")
	go StartMockBackend(":8083")

	// starting a new goroutine for backend health check
	go BackendHealthCheck(NewServerPool, time.Duration(Config.BackendHealthCheckInterval)*time.Second)

	// starting Proxy server
	mux := http.NewServeMux()
	// handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// creating and forwarding of request is handle by reverseproxy
		backend := NewServerPool.GetNextBackend()
		backend.ReverseProxy.ServeHTTP(w, r)

	})
	// creating server
	proxyServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", Config.Port),
		Handler: mux,
	}

	fmt.Println("Proxy Server is listening on port:", Config.Port)
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
