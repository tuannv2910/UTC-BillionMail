package llm

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Built-in provider names.
const (
	ProviderOpenAI           = "openai"
	ProviderAnthropic        = "anthropic"
	ProviderGemini           = "gemini"
	ProviderOpenAICompatible = "openai_compatible"
)

// Factory constructs a Provider from its config.
type Factory func(cfg Config) (Provider, error)

// Registry maps provider names to factories. New vendors are added with Register,
// so callers never switch on vendor names themselves.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds or replaces the factory for name (case-insensitive).
func (r *Registry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[normalizeName(name)] = factory
}

// New builds the provider named by cfg.Provider.
func (r *Registry) New(cfg Config) (Provider, error) {
	r.mu.RLock()
	factory, ok := r.factories[normalizeName(cfg.Provider)]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, cfg.Provider)
	}
	return factory(cfg)
}

// Names lists the registered provider names in sorted order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// defaultRegistry holds the built-in providers.
var defaultRegistry = func() *Registry {
	r := NewRegistry()
	r.Register(ProviderOpenAI, NewOpenAI)
	r.Register(ProviderAnthropic, NewAnthropic)
	r.Register(ProviderGemini, NewGemini)
	r.Register(ProviderOpenAICompatible, NewOpenAICompatible)
	return r
}()

// Register adds a provider factory to the default registry.
func Register(name string, factory Factory) { defaultRegistry.Register(name, factory) }

// New builds a provider from the default registry.
func New(cfg Config) (Provider, error) { return defaultRegistry.New(cfg) }

// Names lists the providers in the default registry.
func Names() []string { return defaultRegistry.Names() }
