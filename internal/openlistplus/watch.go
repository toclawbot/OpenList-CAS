package openlistplus

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
)

type watchState struct {
	cancel context.CancelFunc
}

var (
	watchedStorages sync.Map
	inFlightRestore sync.Map
)

func StartAutoRestoreExistingCAS(ctx context.Context, storage driver.Driver) error {
	StopAutoRestoreExistingCAS(storage)
	if !ShouldAutoRestore(storage) {
		return nil
	}
	watchCtx, cancel := context.WithCancel(context.Background())
	watchedStorages.Store(watchedStorageKey(storage), &watchState{cancel: cancel})
	go autoRestoreLoop(watchCtx, storage)
	return nil
}

func StopAutoRestoreExistingCAS(storage driver.Driver) {
	key := watchedStorageKey(storage)
	if stateValue, ok := watchedStorages.LoadAndDelete(key); ok {
		if state, ok := stateValue.(*watchState); ok && state.cancel != nil {
			state.cancel()
		}
	}
	clearRestorePrefix(key + ":")
}

func HandleObjsUpdate(ctx context.Context, parent string, objs []model.Obj) {
	storage, actualParent, err := op.GetStorageAndActualPath(parent)
	if err != nil || !ShouldAutoRestore(storage) || !ShouldAutoRestorePath(storage, actualParent) {
		return
	}
	for _, obj := range objs {
		if obj == nil || obj.IsDir() || !HasCASSuffix(obj.GetName()) {
			continue
		}
		casPath := path.Join(actualParent, obj.GetName())
		if !beginAutoRestore(storage, casPath) {
			continue
		}
		func() {
			defer endAutoRestore(storage, casPath)
			_ = restoreCASPath(ctx, storage, actualParent, casPath)
		}()
	}
}

func watchedStorageKey(storage driver.Driver) string {
	return fmt.Sprintf("%s:%p", storage.Config().Name, storage)
}

func autoRestoreLoop(ctx context.Context, storage driver.Driver) {
	_ = scanConfiguredCASPaths(ctx, storage)
	for {
		wait := autoRestoreInterval(storage)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			_ = scanConfiguredCASPaths(ctx, storage)
		}
	}
}

func autoRestoreInterval(storage driver.Driver) time.Duration {
	ttl := storage.GetStorage().CacheExpiration
	if ttl <= 0 {
		ttl = 30
	}
	return time.Minute * time.Duration(ttl)
}

func scanConfiguredCASPaths(ctx context.Context, storage driver.Driver) error {
	paths := AutoRestorePaths(storage)
	if len(paths) == 0 {
		return ctx.Err()
	}
	for _, actualPath := range paths {
		if err := scanCASDir(ctx, storage, actualPath); err != nil && ctx.Err() == nil {
			continue
		}
	}
	return ctx.Err()
}

func scanCASDir(ctx context.Context, storage driver.Driver, actualPath string) error {
	if utils.IsCanceled(ctx) {
		return ctx.Err()
	}
	objs, err := op.List(ctx, storage, actualPath, model.ListArgs{
		Refresh:  true,
		SkipHook: true,
	})
	if err != nil {
		return err
	}
	for _, obj := range objs {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}
		childPath := path.Join(actualPath, obj.GetName())
		if obj.IsDir() {
			_ = scanCASDir(ctx, storage, childPath)
			continue
		}
		if !HasCASSuffix(obj.GetName()) {
			continue
		}
		if !beginAutoRestore(storage, childPath) {
			continue
		}
		func() {
			defer endAutoRestore(storage, childPath)
			_ = restoreCASPath(ctx, storage, actualPath, childPath)
		}()
	}
	return nil
}

func restoreCASPath(ctx context.Context, storage driver.Driver, dirPath, casPath string) error {
	dstDir, err := op.GetUnwrap(ctx, storage, dirPath)
	if err != nil {
		return err
	}
	casObj, err := op.GetUnwrap(ctx, storage, casPath)
	if err != nil {
		return err
	}
	link, err := storage.Link(ctx, casObj, model.LinkArgs{})
	if err != nil {
		return err
	}
	fs := &stream.FileStream{
		Ctx:      ctx,
		Obj:      casObj,
		Mimetype: "application/octet-stream",
	}
	ss, err := stream.NewSeekableStream(fs, link)
	if err != nil {
		link.Close()
		return err
	}
	defer ss.Close()
	info, err := ParseCASFile(ss)
	if err != nil {
		return err
	}
	if _, err = RestoreFromCAS(ctx, storage, dstDir, casObj.GetName(), info); err != nil {
		return err
	}
	if ShouldDeleteCASAfterRestore(storage) {
		return op.Remove(ctx, storage, casPath)
	}
	return nil
}

func beginAutoRestore(storage driver.Driver, casPath string) bool {
	key := watchedStorageKey(storage) + ":" + casPath
	_, loaded := inFlightRestore.LoadOrStore(key, struct{}{})
	return !loaded
}

func endAutoRestore(storage driver.Driver, casPath string) {
	key := watchedStorageKey(storage) + ":" + casPath
	inFlightRestore.Delete(key)
}

func clearRestorePrefix(prefix string) {
	inFlightRestore.Range(func(key, _ any) bool {
		s, ok := key.(string)
		if ok && strings.HasPrefix(s, prefix) {
			inFlightRestore.Delete(s)
		}
		return true
	})
}
