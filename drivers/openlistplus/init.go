package openlistplus

import (
	"context"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	core "github.com/OpenListTeam/OpenList/v4/internal/openlistplus"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
)

func init() {
	core.RegisterHook()
	register := func(storageName string, skipPrepareCAS bool) {
		core.RegisterHandler(core.Handler{
			StorageName:    storageName,
			SkipPrepareCAS: skipPrepareCAS,
			ChunkSize: func(storage driver.Driver, size int64) int64 {
				bridge, ok := storage.(core.Bridge)
				if !ok {
					return 0
				}
				return bridge.OpenListPlusChunkSize(size)
			},
			WriteCAS: func(ctx context.Context, storage driver.Driver, dstDir model.Obj, info *casfile.Info) (model.Obj, error) {
				return storage.(core.Bridge).OpenListPlusWriteCAS(ctx, dstDir, info)
			},
			DeleteSource: func(ctx context.Context, storage driver.Driver, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error {
				return storage.(core.Bridge).OpenListPlusDeleteSourceAfterCAS(ctx, dstDir, uploadedObj, sourceName)
			},
			RestoreFromCAS: func(ctx context.Context, storage driver.Driver, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error) {
				return storage.(core.Bridge).OpenListPlusRestoreFromCAS(ctx, dstDir, casFileName, info)
			},
			StartAutoRestore: func(ctx context.Context, storage driver.Driver) error {
				return storage.(core.Bridge).OpenListPlusStartAutoRestore(ctx)
			},
			StopAutoRestore: func(storage driver.Driver) {
				storage.(core.Bridge).OpenListPlusStopAutoRestore()
			},
			HandleObjsUpdate: func(ctx context.Context, storage driver.Driver, parent string, objs []model.Obj) {
				storage.(core.Bridge).OpenListPlusHandleObjsUpdate(ctx, parent, objs)
			},
		})
	}
	register("189Cloud", true)
	register("189CloudPC", true)
	register("Local", false)
}
