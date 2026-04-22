package openlistplus

import (
	"context"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
)

type Addition struct {
	GenerateCAS                 bool   `json:"generate_cas" help:"Generate a .cas file after uploading a normal file"`
	DeleteSource                bool   `json:"delete_source" help:"Delete the source file after generating the .cas file"`
	RestoreSourceFromCAS        bool   `json:"restore_source_from_cas" help:"Automatically restore the source file when handling a .cas file"`
	RestoreSourceUseCurrentName bool   `json:"restore_source_use_current_name" help:"When restoring from a .cas file, use the current .cas filename and keep the original file extension"`
	DeleteCASAfterRestore       bool   `json:"delete_cas_after_restore" help:"Delete the .cas file after the source file is restored successfully"`
	AutoRestoreExistingCAS      bool   `json:"auto_restore_existing_cas" help:"Automatically scan monitored directories and restore existing .cas files"`
	AutoRestoreExistingCASPaths string `json:"auto_restore_existing_cas_paths" help:"Monitored directories for automatic .cas restore, one path per line; leave empty to disable monitoring"`
}

type AdditionProvider interface {
	OpenListPlusAddition() *Addition
}

type Bridge interface {
	OpenListPlusChunkSize(size int64) int64
	OpenListPlusWriteCAS(ctx context.Context, dstDir model.Obj, info *casfile.Info) (model.Obj, error)
	OpenListPlusDeleteSourceAfterCAS(ctx context.Context, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error
	OpenListPlusRestoreFromCAS(ctx context.Context, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error)
	OpenListPlusStartAutoRestore(ctx context.Context) error
	OpenListPlusStopAutoRestore()
	OpenListPlusHandleObjsUpdate(ctx context.Context, parent string, objs []model.Obj)
}

type Handler struct {
	StorageName      string
	SkipPrepareCAS   bool
	ChunkSize        func(storage driver.Driver, size int64) int64
	WriteCAS         func(ctx context.Context, storage driver.Driver, dstDir model.Obj, info *casfile.Info) (model.Obj, error)
	DeleteSource     func(ctx context.Context, storage driver.Driver, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error
	RestoreFromCAS   func(ctx context.Context, storage driver.Driver, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error)
	StartAutoRestore func(ctx context.Context, storage driver.Driver) error
	StopAutoRestore  func(storage driver.Driver)
	HandleObjsUpdate func(ctx context.Context, storage driver.Driver, parent string, objs []model.Obj)
}

type PreparedPut struct {
	Handled bool
	Obj     model.Obj
	Stream  model.FileStreamer
	CAS     *casfile.Info
}
