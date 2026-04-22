package openlistplus

import (
	"context"
	"encoding/hex"
	"io"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
)

func PreparePut(ctx context.Context, storage driver.Driver, dstDir model.Obj, stream model.FileStreamer) (*PreparedPut, error) {
	prepared := &PreparedPut{Stream: stream}
	if ShouldRestoreCAS(storage, stream.GetName()) {
		info, err := ParseCASFile(stream)
		if err != nil {
			return nil, err
		}
		obj, err := RestoreFromCAS(ctx, storage, dstDir, stream.GetName(), info)
		if err != nil {
			return nil, err
		}
		prepared.Handled = true
		prepared.Obj = obj
		return prepared, nil
	}
	if !ShouldGenerateCAS(storage, stream.GetName()) {
		return prepared, nil
	}
	handler, ok := handlerFor(storage)
	if !ok {
		return prepared, nil
	}
	if handler.SkipPrepareCAS {
		return prepared, nil
	}
	chunkSize := int64(0)
	if handler.ChunkSize != nil {
		chunkSize = handler.ChunkSize(storage, stream.GetSize())
	}
	info, err := buildCASInfo(stream, chunkSize)
	if err != nil {
		return nil, err
	}
	prepared.CAS = info
	return prepared, nil
}

func FinishPut(ctx context.Context, storage driver.Driver, dstDir model.Obj, prepared *PreparedPut, uploadedObj model.Obj) (model.Obj, error) {
	if prepared == nil || prepared.CAS == nil {
		return uploadedObj, nil
	}
	handler, ok := handlerFor(storage)
	if !ok || handler.WriteCAS == nil {
		return uploadedObj, nil
	}
	casObj, err := handler.WriteCAS(ctx, storage, dstDir, prepared.CAS)
	if err != nil {
		return nil, err
	}
	if ShouldDeleteSource(storage) && handler.DeleteSource != nil {
		if err = handler.DeleteSource(ctx, storage, dstDir, uploadedObj, prepared.CAS.Name); err != nil {
			return nil, err
		}
		return casObj, nil
	}
	if uploadedObj != nil {
		return uploadedObj, nil
	}
	return casObj, nil
}

func buildCASInfo(stream model.FileStreamer, chunkSize int64) (*casfile.Info, error) {
	cache, err := stream.CacheFullAndWriter(nil, nil)
	if err != nil {
		return nil, err
	}
	size := stream.GetSize()
	if size < 0 {
		if size, err = cache.Seek(0, io.SeekEnd); err != nil {
			return nil, err
		}
	}
	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	fullHasher := utils.MD5.NewFunc()
	partHasher := utils.MD5.NewFunc()
	partMD5s := make([]string, 0)
	if chunkSize <= 0 {
		chunkSize = size
	}
	remaining := size
	buf := make([]byte, minInt64(chunkSize, 1024*1024))
	for remaining > 0 {
		partRemaining := minInt64(chunkSize, remaining)
		partHasher = utils.MD5.NewFunc()
		for partRemaining > 0 {
			readSize := int(minInt64(int64(len(buf)), partRemaining))
			n, readErr := io.ReadFull(cache, buf[:readSize])
			if n > 0 {
				fullHasher.Write(buf[:n])
				partHasher.Write(buf[:n])
				partRemaining -= int64(n)
				remaining -= int64(n)
			}
			if readErr != nil {
				if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
					break
				}
				return nil, readErr
			}
		}
		partMD5s = append(partMD5s, strings.ToUpper(hex.EncodeToString(partHasher.Sum(nil))))
	}
	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	fileMD5 := hex.EncodeToString(fullHasher.Sum(nil))
	sliceMD5 := fileMD5
	if len(partMD5s) > 1 {
		sliceMD5 = utils.GetMD5EncodeStr(strings.Join(partMD5s, "\n"))
	}
	return casfile.New(stream.GetName(), size, fileMD5, sliceMD5), nil
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
