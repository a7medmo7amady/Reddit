package consul

import (
	"fmt"
	"math/rand"

	consulapi "github.com/hashicorp/consul/api"
)

// Resolver resolves a service name to a healthy instance URL via Consul.
type Resolver struct {
	client *consulapi.Client
}

func New(addr string) (*Resolver, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = addr
	client, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}
	return &Resolver{client: client}, nil
}

// Resolve returns a random healthy instance URL ("http://host:port") for the
// given Consul service name.
func (r *Resolver) Resolve(service string) (string, error) {
	entries, _, err := r.client.Health().Service(service, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("consul lookup %q: %w", service, err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instances of %q in consul", service)
	}
	e := entries[rand.Intn(len(entries))]
	host := e.Service.Address
	if host == "" {
		host = e.Node.Address
	}
	return fmt.Sprintf("http://%s:%d", host, e.Service.Port), nil
}
