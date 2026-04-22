package openlistplus

import "github.com/OpenListTeam/OpenList/v4/internal/driver"

func handlerFor(storage driver.Driver) (Handler, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	handler, ok := registry[storage.Config().Name]
	return handler, ok
}
