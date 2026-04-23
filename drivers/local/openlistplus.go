package local

import (
	"context"
	"os"
	"path/filepath"

	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
)

func (d *Local) OpenListPlusChunkSize(size int64) int64 {
	return 10 * 1024 * 1024
}

func (d *Local) OpenListPlusWriteCAS(ctx context.Context, dstDir model.Obj, info *casfile.Info) (model.Obj, error) {
	body, err := casfile.MarshalBase64(info)
	if err != nil {
		return nil, err
	}
	casName := info.Name + ".cas"
	casPath := filepath.Join(dstDir.GetPath(), casName)
	if err = os.WriteFile(casPath, []byte(body), 0o644); err != nil {
		return nil, err
	}
	if d.directoryMap.Has(dstDir.GetPath()) {
		d.directoryMap.UpdateDirSize(dstDir.GetPath())
		d.directoryMap.UpdateDirParents(dstDir.GetPath())
	}
	return &model.Object{
		Path: casPath,
		Name: casName,
		Size: int64(len(body)),
	}, nil
}

func (d *Local) OpenListPlusDeleteSourceAfterCAS(ctx context.Context, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error {
	fullPath := filepath.Join(dstDir.GetPath(), sourceName)
	if err := os.Remove(fullPath); err != nil {
		return err
	}
	if d.directoryMap.Has(dstDir.GetPath()) {
		d.directoryMap.UpdateDirSize(dstDir.GetPath())
		d.directoryMap.UpdateDirParents(dstDir.GetPath())
	}
	return nil
}

func (d *Local) OpenListPlusDeletePermanently(ctx context.Context, obj model.Obj) error {
	if obj == nil {
		return nil
	}
	fullPath := obj.GetPath()
	if fullPath == "" {
		return errs.NotImplement
	}
	if err := os.Remove(fullPath); err != nil {
		return err
	}
	dirPath := filepath.Dir(fullPath)
	if d.directoryMap.Has(dirPath) {
		d.directoryMap.UpdateDirSize(dirPath)
		d.directoryMap.UpdateDirParents(dirPath)
	}
	return nil
}

func (d *Local) OpenListPlusRestoreFromCAS(ctx context.Context, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error) {
	return nil, errs.NotImplement
}

func (d *Local) OpenListPlusStartAutoRestore(ctx context.Context) error {
	return nil
}

func (d *Local) OpenListPlusStopAutoRestore() {}

func (d *Local) OpenListPlusHandleObjsUpdate(ctx context.Context, parent string, objs []model.Obj) {}
