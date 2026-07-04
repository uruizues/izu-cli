package provider

import "fmt"

type Registry struct {
    providers map[string]Provider
    current   string
}

func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]Provider),
    }
}

func (r *Registry) Register(p Provider) {
    r.providers[p.Name()] = p
}

func (r *Registry) SetCurrent(name string) error {
    if _, ok := r.providers[name]; !ok {
        return fmt.Errorf("provider %s not found", name)
    }
    r.current = name
    return nil
}

func (r *Registry) Current() Provider {
    return r.providers[r.current]
}

func (r *Registry) List() []string {
    var names []string
    for name := range r.providers {
        names = append(names, name)
    }
    return names
}

func (r *Registry) Get(name string) (Provider, bool) {
    p, ok := r.providers[name]
    return p, ok
}
