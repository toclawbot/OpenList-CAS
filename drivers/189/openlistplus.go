package _189

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/drivers/base"
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

const openListPlus189SliceSize int64 = 10 * 1024 * 1024

func (d *Cloud189) OpenListPlusChunkSize(size int64) int64 {
	return openListPlus189SliceSize
}

func (d *Cloud189) OpenListPlusWriteCAS(ctx context.Context, dstDir model.Obj, info *casfile.Info) (model.Obj, error) {
	body, err := casfile.MarshalBase64(info)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	name := openlistplus.BuildCASName(info.Name)
	file := &stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     int64(len(body)),
			Modified: now,
			Ctime:    now,
		},
		Reader:   strings.NewReader(body),
		Mimetype: "text/plain",
	}
	if err = d.newUpload(ctx, dstDir, file, nil); err != nil {
		return nil, err
	}
	return &model.Object{Name: name, Size: int64(len(body)), Modified: now, Ctime: now}, nil
}

func (d *Cloud189) OpenListPlusDeleteSourceAfterCAS(ctx context.Context, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error {
	if uploadedObj != nil && uploadedObj.GetID() != "" {
		return d.Remove(ctx, uploadedObj)
	}
	obj, err := d.findOpenListPlusFileByName(dstDir.GetID(), sourceName)
	if err != nil {
		return err
	}
	return d.Remove(ctx, obj)
}

func (d *Cloud189) OpenListPlusRestoreFromCAS(ctx context.Context, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error) {
	useCurrentName := openlistplus.ShouldUseCurrentRestoreName(d) || openlistplus.HasPreviewRestorePrefix(casFileName)
	sourceName := openlistplus.ResolveRestoreName(casFileName, info, useCurrentName)
	sessionKey, err := d.getSessionKey()
	if err != nil {
		return nil, err
	}
	d.sessionKey = sessionKey

	initResp, err := d.uploadRequest("/person/initMultiUpload", map[string]string{
		"parentFolderId": dstDir.GetID(),
		"fileName":       encode(sourceName),
		"fileSize":       strconv.FormatInt(info.Size, 10),
		"sliceSize":      strconv.FormatInt(openListPlus189SliceSize, 10),
		"lazyCheck":      "1",
		"fileMd5":        strings.ToLower(info.MD5),
		"sliceMd5":       strings.ToLower(info.SliceMD5),
	}, nil)
	if err != nil {
		return nil, err
	}
	uploadFileID := jsoniter.Get(initResp, "data", "uploadFileId").ToString()
	if uploadFileID == "" {
		return nil, fmt.Errorf("189 restore: missing uploadFileId")
	}

	fileExists := jsoniter.Get(initResp, "data", "fileDataExists").ToInt()
	if fileExists != 1 {
		if _, err = d.uploadRequest("/person/checkTransSecond", map[string]string{
			"uploadFileId": uploadFileID,
			"fileMd5":      strings.ToLower(info.MD5),
			"sliceMd5":     strings.ToLower(info.SliceMD5),
		}, nil); err != nil {
			return nil, err
		}
	}

	if _, err = d.uploadRequest("/person/commitMultiUploadFile", map[string]string{
		"uploadFileId": uploadFileID,
		"fileMd5":      strings.ToLower(info.MD5),
		"sliceMd5":     strings.ToLower(info.SliceMD5),
		"lazyCheck":    "1",
		"opertype":     "3",
	}, nil); err != nil {
		return nil, err
	}
	if obj, findErr := d.findOpenListPlusFileByName(dstDir.GetID(), sourceName); findErr == nil {
		return obj, nil
	}
	return &model.Object{
		Name:     sourceName,
		Size:     info.Size,
		Modified: time.Now(),
		Ctime:    time.Now(),
		HashInfo: utils.NewHashInfo(utils.MD5, info.MD5),
	}, nil
}

func (d *Cloud189) OpenListPlusStartAutoRestore(ctx context.Context) error {
	return nil
}

func (d *Cloud189) OpenListPlusStopAutoRestore() {}

func (d *Cloud189) OpenListPlusHandleObjsUpdate(ctx context.Context, parent string, objs []model.Obj) {}

func (d *Cloud189) OpenListPlusPreviewTarget(ctx context.Context, casFileName string, info *casfile.Info) (driver.Driver, model.Obj, string, string, error) {
	if err := op.MakeDir(ctx, d, "/TEMP"); err != nil {
		return nil, nil, "", "", err
	}
	dstDir, err := op.GetUnwrap(ctx, d, "/TEMP")
	if err != nil {
		return nil, nil, "", "", err
	}
	restoredName := openlistplus.BuildPreviewRestoreName(casFileName, info, openlistplus.ShouldUseCurrentRestoreName(d))
	return d, dstDir, restoredName, path.Join("/TEMP", restoredName), nil
}

func (d *Cloud189) OpenListPlusDeletePreviewRestoredPermanently(ctx context.Context, obj model.Obj) error {
	isFolder := 0
	if obj.IsDir() {
		isFolder = 1
	}
	taskInfos := []base.Json{
		{
			"fileId":   obj.GetID(),
			"fileName": obj.GetName(),
			"isFolder": isFolder,
		},
	}
	taskInfosBytes, err := utils.Json.Marshal(taskInfos)
	if err != nil {
		return err
	}
	form := map[string]string{
		"type":           "DELETE",
		"targetFolderId": "",
		"taskInfos":      string(taskInfosBytes),
	}
	if _, err = d.request("https://cloud.189.cn/api/open/batch/createBatchTask.action", http.MethodPost, func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(form)
	}, nil); err != nil {
		return err
	}
	form["type"] = "CLEAR_RECYCLE"
	_, err = d.request("https://cloud.189.cn/api/open/batch/createBatchTask.action", http.MethodPost, func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(form)
	}, nil)
	return err
}

func (d *Cloud189) findOpenListPlusFileByName(folderID, name string) (model.Obj, error) {
	files, err := d.getFiles(folderID)
	if err != nil {
		return nil, err
	}
	for _, obj := range files {
		if obj.GetName() == name {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("189 file not found: %s", name)
}
