package openlistplus

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
)

const previewRestorePrefix = "TEMP_"

func BuildCASName(sourceName string) string {
	return sourceName + ".cas"
}

func HasCASSuffix(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".cas")
}

func TrimCASSuffix(name string) string {
	if !HasCASSuffix(name) {
		return name
	}
	return name[:len(name)-4]
}

func HasPreviewRestorePrefix(name string) bool {
	return strings.HasPrefix(TrimCASSuffix(name), previewRestorePrefix)
}

func BuildPreviewRestoreName(casFileName string, info *casfile.Info, useCurrentName bool) string {
	return fmt.Sprintf("%s%05d_%s", previewRestorePrefix, randomPreviewRestoreCode(), ResolveRestoreName(casFileName, info, useCurrentName))
}

func BuildPreviewRestoreCASName(restoredName string) string {
	return restoredName + ".cas"
}

func ResolveRestoreName(casFileName string, info *casfile.Info, useCurrentName bool) string {
	if info == nil {
		return TrimCASSuffix(casFileName)
	}
	if !useCurrentName {
		return info.Name
	}
	currentName := TrimCASSuffix(casFileName)
	if currentName == "" {
		return info.Name
	}
	sourceExt := normalizedSourceExtension(info.Name)
	if sourceExt == "" {
		return currentName
	}
	baseName := strings.TrimSuffix(currentName, path.Ext(currentName))
	if baseName == "" {
		return currentName
	}
	return baseName + sourceExt
}

func normalizedSourceExtension(name string) string {
	ext := strings.ToLower(path.Ext(name))
	if len(ext) <= 1 {
		return ""
	}
	return ext
}

func randomPreviewRestoreCode() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err == nil {
		return n.Int64()
	}
	return time.Now().UnixNano() % 100000
}
