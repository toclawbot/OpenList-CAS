package _189pc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/go-resty/resty/v2"
)

func (y *Cloud189PC) OpenListPlusChunkSize(size int64) int64 {
	return partSize(size)
}

func (y *Cloud189PC) OpenListPlusWriteCAS(ctx context.Context, dstDir model.Obj, info *casfile.Info) (model.Obj, error) {
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

	if !y.isFamily() && y.FamilyTransfer {
		familyObj, _, err := y.StreamUpload(ctx, y.familyTransferFolder, file, nil, true, false)
		if err != nil {
			return nil, err
		}
		if familyObj == nil {
			familyObj = &model.Object{Name: name, Size: int64(len(body)), Modified: now, Ctime: now}
		}
		if err = y.SaveFamilyFileToPersonCloud(ctx, y.FamilyID, familyObj, dstDir, true); err != nil {
			return nil, err
		}
		go y.Delete(context.TODO(), y.FamilyID, familyObj)
		if y.cleanFamilyTransferFile != nil {
			go y.cleanFamilyTransferFile()
		}
		obj, findErr := y.findFileByName(ctx, name, dstDir.GetID(), false)
		if findErr == nil {
			return obj, nil
		}
		return &model.Object{Name: name, Size: int64(len(body)), Modified: now, Ctime: now}, nil
	}

	obj, _, err := y.StreamUpload(ctx, dstDir, file, nil, y.isFamily(), true)
	if err != nil {
		return nil, err
	}
	if obj != nil {
		return obj, nil
	}
	return &model.Object{Name: name, Size: int64(len(body)), Modified: now, Ctime: now}, nil
}

func (y *Cloud189PC) OpenListPlusDeleteSourceAfterCAS(ctx context.Context, dstDir model.Obj, uploadedObj model.Obj, sourceName string) error {
	if uploadedObj != nil {
		return y.OpenListPlusDeletePermanently(ctx, uploadedObj)
	}
	obj, err := y.findFileByName(ctx, sourceName, dstDir.GetID(), y.isFamily())
	if err != nil {
		return err
	}
	return y.OpenListPlusDeletePermanently(ctx, obj)
}

func (y *Cloud189PC) OpenListPlusDeletePermanently(ctx context.Context, obj model.Obj) error {
	if obj == nil {
		return nil
	}
	return y.Delete(ctx, IF(y.isFamily(), y.FamilyID, ""), model.UnwrapObjName(obj))
}

func (y *Cloud189PC) OpenListPlusRestoreFromCAS(ctx context.Context, dstDir model.Obj, casFileName string, info *casfile.Info) (model.Obj, error) {
	useCurrentName := openlistplus.ShouldUseCurrentRestoreName(y) || openlistplus.HasPreviewRestorePrefix(casFileName)
	sourceName := openlistplus.ResolveRestoreName(casFileName, info, useCurrentName)
	isFamily := y.isFamily()
	fullURL := UPLOAD_URL
	if isFamily {
		fullURL += "/family"
	} else {
		fullURL += "/person"
	}

	params := Params{
		"parentFolderId": dstDir.GetID(),
		"fileName":       url.QueryEscape(sourceName),
		"fileSize":       strconv.FormatInt(info.Size, 10),
		"fileMd5":        strings.ToUpper(info.MD5),
		"sliceMd5":       strings.ToUpper(info.SliceMD5),
		"sliceSize":      strconv.FormatInt(partSize(info.Size), 10),
	}
	if isFamily {
		params.Set("familyId", y.FamilyID)
	}
	var initResp InitMultiUploadResp
	if _, err := y.request(fullURL+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, params, &initResp, isFamily); err != nil {
		return nil, err
	}
	if initResp.Data.UploadFileID == "" {
		return nil, fmt.Errorf("189pc restore: missing uploadFileId")
	}
	if initResp.Data.FileDataExists != 1 {
		return nil, fmt.Errorf("189pc restore: file data not found on remote side")
	}

	var commitResp CommitMultiUploadFileResp
	if _, err := y.request(fullURL+"/commitMultiUploadFile", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, Params{
		"uploadFileId": initResp.Data.UploadFileID,
		"isLog":        "0",
		"opertype":     "3",
	}, &commitResp, isFamily); err != nil {
		return nil, err
	}
	if obj := commitResp.toFile(); obj != nil {
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

func (y *Cloud189PC) OpenListPlusStartAutoRestore(ctx context.Context) error {
	return nil
}

func (y *Cloud189PC) OpenListPlusStopAutoRestore() {}

func (y *Cloud189PC) OpenListPlusHandleObjsUpdate(ctx context.Context, parent string, objs []model.Obj) {}

func (y *Cloud189PC) OpenListPlusPreviewTarget(ctx context.Context, casFileName string, info *casfile.Info) (driver.Driver, model.Obj, string, string, error) {
	target := y.openListPlusPreviewStorage()
	if err := op.MakeDir(ctx, target, "/TEMP"); err != nil {
		return nil, nil, "", "", err
	}
	dstDir, err := op.GetUnwrap(ctx, target, "/TEMP")
	if err != nil {
		return nil, nil, "", "", err
	}
	restoredName := openlistplus.BuildPreviewRestoreName(casFileName, info, openlistplus.ShouldUseCurrentRestoreName(y))
	return target, dstDir, restoredName, path.Join("/TEMP", restoredName), nil
}

func (y *Cloud189PC) OpenListPlusDeletePreviewRestoredPermanently(ctx context.Context, obj model.Obj) error {
	return y.OpenListPlusDeletePermanently(ctx, obj)
}

func (y *Cloud189PC) openListPlusPreviewStorage() *Cloud189PC {
	target := *y
	target.ref = y
	target.cron = nil
	if y.isFamily() || y.FamilyTransfer {
		target.Type = "family"
		target.FamilyID = y.FamilyID
		target.RootFolderID = ""
		return &target
	}
	target.Type = "personal"
	target.FamilyID = ""
	target.RootFolderID = "-11"
	return &target
}
