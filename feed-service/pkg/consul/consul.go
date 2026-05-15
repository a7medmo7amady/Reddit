package consul

import (
	"fmt"
	"log"
	"math/rand"

	consulapi "github.com/hashicorp/consul/api"
)

type Resolver struct {
	client   *consulapi.Client
	fallback map[string]string
}

// New creates a Consul resolver. Falls back gracefully if Consul is unreachable.
func New(addr string, fallback map[string]string) *Resolver {
	cfg := consulapi.DefaultConfig()
	cfg.Address = addr
	client, err := consulapi.NewClient(cfg)
	if err != nil {
		log.Printf("[Consul] client init failed (%v), using static fallbacks", err)
		return &Resolver{fallback: fallback}
	}
	return &Resolver{client: client, fallback: fallback}
}

// Resolve returns a random healthy instance URL ("http://host:port") for the
// given service name. Falls back to the static map if Consul lookup fails.
func (r *Resolver) Resolve(service string) string {
	if r.client != nil {
		entries, _, err := r.client.Health().Service(service, "", true, nil)
		if err == nil && len(entries) > 0 {
			e := entries[rand.Intn(len(entries))]
			host := e.Service.Address
			if host == "" {
				host = e.Node.Address
			}
			url := fmt.Sprintf("http://%s:%d", host, e.Service.Port)
			log.Printf("[Consul] resolved %s → %s", service, url)
			return url
		}
		log.Printf("[Consul] lookup %q failed (%v), falling back to static", service, err)
	}
	if url, ok := r.fallback[service]; ok {
		return url
	}
	return ""
}
