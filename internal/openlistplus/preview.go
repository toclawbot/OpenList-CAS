package openlistplus

import (
	"context"
	stdpath "path"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
)

type previewTargetResolver interface {
	OpenListPlusPreviewTarget(ctx context.Context, casFileName string, info *casfile.Info) (driver.Driver, model.Obj, string, string, error)
}

func CanPreviewCAS(storage driver.Driver, name string) bool {
	if storage == nil || !HasCASSuffix(name) {
		return false
	}
	switch storage.Config().Name {
	case "189Cloud", "189CloudPC":
		return true
	default:
		return false
	}
}

func ResolveCASPreviewNameByMountPath(ctx context.Context, mountPath string) (string, error) {
	storage, actualPath, err := op.GetStorageAndActualPath(mountPath)
	if err != nil {
		return "", err
	}
	obj, err := op.Get(ctx, storage, actualPath)
	if err != nil {
		return "", err
	}
	return ResolveCASPreviewName(ctx, storage, obj)
}

func ResolveCASPreviewName(ctx context.Context, storage driver.Driver, obj model.Obj) (string, error) {
	if !CanPreviewCAS(storage, obj.GetName()) {
		return obj.GetName(), nil
	}
	info, err := readCASInfo(ctx, storage, obj)
	if err != nil {
		return "", err
	}
	return ResolveRestoreName(obj.GetName(), info, ShouldUseCurrentRestoreName(storage)), nil
}

func ResolveCASPreviewLinkByMountPath(ctx context.Context, mountPath string, args model.LinkArgs) (*model.Link, model.Obj, string, error) {
	storage, actualPath, err := op.GetStorageAndActualPath(mountPath)
	if err != nil {
		return nil, nil, "", err
	}
	obj, err := op.Get(ctx, storage, actualPath)
	if err != nil {
		return nil, nil, "", err
	}
	if !CanPreviewCAS(storage, obj.GetName()) {
		return nil, nil, "", nil
	}
	info, err := readCASInfo(ctx, storage, obj)
	if err != nil {
		return nil, nil, "", err
	}
	targetStorage, dstDir, restoredName, restoredPath, err := resolvePreviewTarget(ctx, storage, actualPath, obj.GetName(), info)
	if err != nil {
		return nil, nil, "", err
	}
	shouldDeleteNow := false
	restoredObj, err := findObjectByName(ctx, targetStorage, dstDir, restoredName)
	switch {
	case err == nil:
		shouldDeleteNow = strings.HasPrefix(restoredName, previewRestorePrefix)
	case errs.IsObjectNotFound(err):
		previewCASName := BuildPreviewRestoreCASName(obj.GetName(), info, ShouldUseCurrentRestoreName(storage))
		if _, err = RestoreFromCAS(ctx, targetStorage, dstDir, previewCASName, info); err != nil {
			return nil, nil, "", err
		}
		restoredObj, err = findObjectByName(ctx, targetStorage, dstDir, restoredName)
		if err != nil {
			return nil, nil, "", err
		}
		shouldDeleteNow = true
	default:
		return nil, nil, "", err
	}
	link, err := targetStorage.Link(ctx, restoredObj, args)
	if err != nil {
		return nil, nil, "", err
	}
	if shouldDeleteNow {
		deletePreviewRestoredObjNow(ctx, targetStorage, restoredObj, restoredPath)
	}
	return link, restoredObj, restoredName, nil
}

func resolvePreviewTarget(ctx context.Context, storage driver.Driver, actualPath, casFileName string, info *casfile.Info) (driver.Driver, model.Obj, string, string, error) {
	if resolver, ok := storage.(previewTargetResolver); ok {
		return resolver.OpenListPlusPreviewTarget(ctx, casFileName, info)
	}
	dirPath := stdpath.Dir(actualPath)
	dstDir, err := op.GetUnwrap(ctx, storage, dirPath)
	if err != nil {
		return nil, nil, "", "", err
	}
	restoredName := BuildPreviewRestoreName(casFileName, info, ShouldUseCurrentRestoreName(storage))
	restoredPath := stdpath.Join(dirPath, restoredName)
	return storage, dstDir, restoredName, restoredPath, nil
}

func readCASInfo(ctx context.Context, storage driver.Driver, obj model.Obj) (*casfile.Info, error) {
	link, err := storage.Link(ctx, obj, model.LinkArgs{})
	if err != nil {
		return nil, err
	}
	fs := &stream.FileStream{
		Ctx:      ctx,
		Obj:      obj,
		Mimetype: "application/octet-stream",
	}
	ss, err := stream.NewSeekableStream(fs, link)
	if err != nil {
		link.Close()
		return nil, err
	}
	defer ss.Close()
	return ParseCASFile(ss)
}

func findObjectByName(ctx context.Context, storage driver.Driver, dir model.Obj, name string) (model.Obj, error) {
	files, err := storage.List(ctx, dir, model.ListArgs{})
	if err != nil {
		return nil, err
	}
	for _, obj := range files {
		if obj.GetName() == name {
			return obj, nil
		}
	}
	return nil, errs.ObjectNotFound
}
