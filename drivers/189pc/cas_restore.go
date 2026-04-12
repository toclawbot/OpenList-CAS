package _189pc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/casfile"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/go-resty/resty/v2"
)

func (y *Cloud189PC) shouldRestoreSourceFromCAS(name string) bool {
	return y.RestoreSourceFromCAS && strings.HasSuffix(strings.ToLower(name), ".cas")
}

func (y *Cloud189PC) restoreSourceFromCAS(ctx context.Context, dstDir model.Obj, file model.FileStreamer) (model.Obj, error) {
	info, err := readCASRestoreInfo(file)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(info.Name, `/\`) {
		return nil, fmt.Errorf("restore from .cas failed: source file name %q contains a path", info.Name)
	}

	isFamily := y.isFamily()
	fullURL := UPLOAD_URL
	if isFamily {
		fullURL += "/family"
	} else {
		fullURL += "/person"
	}

	params := Params{
		"parentFolderId": dstDir.GetID(),
		"fileName":       url.QueryEscape(info.Name),
		"fileSize":       strconv.FormatInt(info.Size, 10),
		"sliceSize":      strconv.FormatInt(partSize(info.Size), 10),
		"fileMd5":        strings.ToUpper(info.MD5),
		"sliceMd5":       strings.ToUpper(info.SliceMD5),
	}
	if isFamily {
		params.Set("familyId", y.FamilyID)
	}

	var initMultiUpload InitMultiUploadResp
	_, err = y.request(fullURL+"/initMultiUpload", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, params, &initMultiUpload, isFamily)
	if err != nil {
		return nil, err
	}
	if initMultiUpload.Data.FileDataExists != 1 {
		return nil, fmt.Errorf("restore from .cas failed: source data for %q was not found in 189CloudPC", info.Name)
	}

	var resp CommitMultiUploadFileResp
	_, err = y.request(fullURL+"/commitMultiUploadFile", http.MethodGet, func(req *resty.Request) {
		req.SetContext(ctx)
	}, Params{
		"uploadFileId": initMultiUpload.Data.UploadFileID,
		"isLog":        "0",
		"opertype":     "3",
	}, &resp, isFamily)
	if err != nil {
		return nil, err
	}
	return resp.toFile(), nil
}

func readCASRestoreInfo(file model.FileStreamer) (*casfile.Info, error) {
	cache, err := file.CacheFullAndWriter(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("cache .cas file: %w", err)
	}

	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek .cas file: %w", err)
	}
	data, err := io.ReadAll(cache)
	if err != nil {
		return nil, fmt.Errorf("read .cas file: %w", err)
	}
	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind .cas file: %w", err)
	}

	info, err := casfile.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse .cas file: %w", err)
	}
	return info, nil
}
