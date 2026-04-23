package openlistplus

import (
	"path"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
)

func GetAddition(storage driver.Driver) *Addition {
	provider, ok := storage.(AdditionProvider)
	if !ok {
		return nil
	}
	addition := provider.OpenListPlusAddition()
	if addition == nil {
		return nil
	}
	addition.Normalize()
	return addition
}

func (a *Addition) Normalize() {
	if a == nil {
		return
	}
}

func ShouldGenerateCAS(storage driver.Driver, name string) bool {
	addition := GetAddition(storage)
	return addition != nil && addition.GenerateCAS && !HasCASSuffix(name)
}

func ShouldDeleteSource(storage driver.Driver) bool {
	addition := GetAddition(storage)
	return addition != nil && addition.DeleteSource
}

func ShouldRestoreCAS(storage driver.Driver, name string) bool {
	addition := GetAddition(storage)
	return addition != nil && addition.RestoreSourceFromCAS && HasCASSuffix(name)
}

func ShouldDeleteCASAfterRestore(storage driver.Driver) bool {
	addition := GetAddition(storage)
	return addition != nil && addition.DeleteCASAfterRestore
}

func ShouldUseCurrentRestoreName(storage driver.Driver) bool {
	return true
}

func ShouldAutoRestore(storage driver.Driver) bool {
	addition := GetAddition(storage)
	return addition != nil && addition.AutoRestoreExistingCAS
}

func AutoRestorePaths(storage driver.Driver) []string {
	addition := GetAddition(storage)
	if addition == nil {
		return nil
	}
	lines := strings.Split(addition.AutoRestoreExistingCASPaths, "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		paths = append(paths, path.Clean("/"+strings.TrimPrefix(line, "/")))
	}
	return paths
}

func ShouldAutoRestorePath(storage driver.Driver, actualPath string) bool {
	paths := AutoRestorePaths(storage)
	if len(paths) == 0 {
		return false
	}
	actualPath = path.Clean("/" + strings.TrimPrefix(actualPath, "/"))
	for _, root := range paths {
		if root == "/" || actualPath == root || strings.HasPrefix(actualPath, root+"/") {
			return true
		}
	}
	return false
}
