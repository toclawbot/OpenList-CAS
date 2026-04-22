package openlistplus

import (
	"context"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type previewPermanentRemover interface {
	OpenListPlusDeletePreviewRestoredPermanently(ctx context.Context, obj model.Obj) error
}

func deletePreviewRestoredNow(storage driver.Driver, actualPath string) {
	if storage == nil || actualPath == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := op.Remove(ctx, storage, actualPath)
	if err != nil && !errs.IsObjectNotFound(err) {
		return
	}
}

func deletePreviewRestoredObjNow(ctx context.Context, storage driver.Driver, obj model.Obj, actualPath string) {
	if storage == nil {
		return
	}
	if obj != nil {
		if permanentRemover, ok := storage.(previewPermanentRemover); ok {
			removeCtx := ctx
			if removeCtx == nil {
				var cancel context.CancelFunc
				removeCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
			}
			err := permanentRemover.OpenListPlusDeletePreviewRestoredPermanently(removeCtx, model.UnwrapObjName(obj))
			if err == nil || errs.IsObjectNotFound(err) {
				return
			}
		}
		if remover, ok := storage.(driver.Remove); ok {
			removeCtx := ctx
			if removeCtx == nil {
				var cancel context.CancelFunc
				removeCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
			}
			err := remover.Remove(removeCtx, model.UnwrapObjName(obj))
			if err == nil || errs.IsObjectNotFound(err) {
				return
			}
		}
	}
	deletePreviewRestoredNow(storage, actualPath)
}
