package server

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var errSubdomainTaken = errors.New("subdomain already taken")

type Tunnel struct {
	ID        string
	Subdomain string
	TargetURL string
	PublicURL string
	CreatedAt time.Time
}

type Registry struct {
	mu          sync.RWMutex
	nextID      int
	bySubdomain map[string]*Tunnel
}

func NewRegistry() *Registry {
	return &Registry{
		bySubdomain: make(map[string]*Tunnel),
	}
}

func (r *Registry) Register(subdomain string, targetURL string, publicURL string) (*Tunnel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bySubdomain[subdomain]; exists {
		return nil, errSubdomainTaken
	}

	r.nextID++
	tunnel := &Tunnel{
		ID:        fmt.Sprintf("tun_%d", r.nextID),
		Subdomain: subdomain,
		TargetURL: targetURL,
		PublicURL: publicURL,
		CreatedAt: time.Now(),
	}

	r.bySubdomain[subdomain] = tunnel
	return tunnel, nil
}

func (r *Registry) Get(subdomain string) (*Tunnel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tunnel, ok := r.bySubdomain[subdomain]
	return tunnel, ok
}
