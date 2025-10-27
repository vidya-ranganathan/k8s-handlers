package registry

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"sync"
)

// Handler interface (re-declared to avoid circular dependency)
type Handler interface {
	Name() string
	Execute(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (interface{}, error)
	Description() string
}

// Registry manages all registered handlers
type Registry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

var (
	globalRegistry = &Registry{
		handlers: make(map[string]Handler),
	}
)

// Register adds a new handler to the global registry
// This is called from init() functions in handler packages
func Register(handler Handler) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	
	name := handler.Name()
	if _, exists := globalRegistry.handlers[name]; exists {
		panic(fmt.Sprintf("handler %s already registered", name))
	}
	
	globalRegistry.handlers[name] = handler
	fmt.Printf("[Registry] Registered handler: %s - %s\n", name, handler.Description())
}

// Get retrieves a handler by name
func Get(name string) (Handler, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	
	handler, exists := globalRegistry.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler %s not found", name)
	}
	
	return handler, nil
}

// List returns all registered handler names and descriptions
func List() map[string]string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	
	result := make(map[string]string)
	for name, handler := range globalRegistry.handlers {
		result[name] = handler.Description()
	}
	
	return result
}

// Execute runs a specific handler by name
func Execute(ctx context.Context, name string, clientset *kubernetes.Clientset, namespace string) (interface{}, error) {
	handler, err := Get(name)
	if err != nil {
		return nil, err
	}
	
	return handler.Execute(ctx, clientset, namespace)
}
