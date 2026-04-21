package openlistplus

import (
	"context"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
)

func RestoreFromCAS(ctx context.Context, storage driver.Driver, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error) {
	handler, ok := handlerFor(storage)
	if !ok || handler.RestoreFromCAS == nil {
		return nil, errs.NotImplement
	}
	return handler.RestoreFromCAS(ctx, storage, dstDir, casFileName, info)
}
