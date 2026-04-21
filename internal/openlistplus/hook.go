package openlistplus

import (
	"sync"

	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

var registerHookOnce sync.Once

func RegisterHook() {
	registerHookOnce.Do(func() {
		op.RegisterObjsUpdateHook(HandleObjsUpdate)
	})
}
