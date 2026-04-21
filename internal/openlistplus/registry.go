package openlistplus

import "sync"

var (
	registryMu sync.RWMutex
	registry   = map[string]Handler{}
)

func RegisterHandler(handler Handler) {
	if handler.StorageName == "" {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[handler.StorageName] = handler
}
