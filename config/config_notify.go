package config

import (
	"fmt"
	"sync"
)

// Observer is the interface that wraps the OnConfigChange methods for both reloadable and non-reloadable configuration changes.
// It is implemented by objects that want to be notified of config changes.
type Observer interface {
	// OnReloadableConfigChange is called when a reloadable configuration key changes.
	OnReloadableConfigChange(key string, oldValue, newValue any)
	// OnNonReloadableConfigChange is called when a non-reloadable configuration key operation happens in the config file.
	OnNonReloadableConfigChange(key string)
}

// NonReloadableConfigChangesFunc is an Observer function invoked for non-reloadable configuration changes.
type NonReloadableConfigChangesFunc func(key string)

func (f NonReloadableConfigChangesFunc) OnNonReloadableConfigChange(key string) {
	f(key)
}

func (f NonReloadableConfigChangesFunc) OnReloadableConfigChange(key string, oldValue, newValue any) {
	// no-op for reloadable changes
}

// ReloadableConfigChangesFunc is an Observer function invoked for reloadable configuration changes.
type ReloadableConfigChangesFunc func(key string, oldValue, newValue any)

func (f ReloadableConfigChangesFunc) OnNonReloadableConfigChange(key string) {
	// no-op for non-reloadable changes
}

func (f ReloadableConfigChangesFunc) OnReloadableConfigChange(key string, oldValue, newValue any) {
	f(key, oldValue, newValue)
}

// notifier manages config change notifications
type notifier struct {
	mu        sync.RWMutex
	observers []Observer
}

// Register registers an observer
func (n *notifier) Register(observer Observer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.observers = append(n.observers, observer)
}

// Unregister removes an observer
func (n *notifier) Unregister(observer Observer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// Find and remove the observer
	for i, obs := range n.observers {
		if obs == observer {
			// Remove the observer by replacing it with the last element
			// and then truncating the slice
			n.observers[i] = n.observers[len(n.observers)-1]
			n.observers = n.observers[:len(n.observers)-1]
			break
		}
	}
}

// notifyNonReloadableConfigChange notifies all observers for non-reloadable configuration changes
func (n *notifier) notifyNonReloadableConfigChange(key string) {
	n.mu.RLock()
	observers := make([]Observer, len(n.observers))
	// Copy the observers to avoid holding the lock while calling observers
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, observer := range observers {
		observer.OnNonReloadableConfigChange(key)
	}
}

// notifyReloadableConfigChange notifies all observers for reloadable configuration changes
func (n *notifier) notifyReloadableConfigChange(key string, oldValue, newValue any) {
	n.mu.RLock()
	observers := make([]Observer, len(n.observers))
	// Copy the observers to avoid holding the lock while calling observers
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, observer := range observers {
		observer.OnReloadableConfigChange(key, oldValue, newValue)
	}
}

// printObserver is a simple observer that prints changes to the console.
// It is enabled by default in the config package and can be used for debugging purposes.
type printObserver struct{}

func (p *printObserver) OnReloadableConfigChange(key string, oldValue, newValue any) {
	fmt.Printf("The value of reloadable key %q changed from %q to %q\n", key, fmt.Sprintf("%+v", oldValue), fmt.Sprintf("%+v", newValue))
}

func (p *printObserver) OnNonReloadableConfigChange(key string) {
	fmt.Printf("Non-reloadable key %q was changed\n", key)
}
