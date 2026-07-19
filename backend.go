package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
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

func IsAliveBackend(rawURL string) bool {
	resp, err := http.Get(rawURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
