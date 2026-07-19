package main

import (
	"sync"
	"time"
)

type ServerPool struct {
	mu       sync.Mutex
	backends []*Backend // url array
	current  int        // current url index
}

func (sp *ServerPool) GetNextBackend() *Backend {
	sp.mu.Lock() // -> why not here read lock ? -> what if 100 req hit at the same milisecond
	defer sp.mu.Unlock()

	// Round Robin Algo Implmenetation
	for i := 0; i < len(sp.backends); i++ {
		idx := (sp.current + i) % len(sp.backends)
		if sp.backends[idx].GetAlive() {
			sp.current = (idx + 1) % len(sp.backends)
			return sp.backends[idx]
		}
	}
	return nil
}

func BackendHealthCheck(sp *ServerPool, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		for _, b := range sp.backends {
			status := IsAliveBackend(b.URL)
			b.SetAlive(status)
		}
	}
}
