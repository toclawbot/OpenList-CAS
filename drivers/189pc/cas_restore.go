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

func (y *Cloud189PC) resolveRestoreSourceName(casFileName string, info *casfile.Info) (string, error) {
	restoreName := info.Name
	if y.RestoreSourceUseCurrentName {
		trimmedName, ok := trimCASSuffix(casFileName)
		if !ok {
			return "", fmt.Errorf("restore from .cas failed: current file name %q does not end with .cas", casFileName)
		}
		restoreName = strings.TrimSpace(trimmedName)
		if restoreName == "" {
			return "", fmt.Errorf("restore from .cas failed: current .cas file name %q has an empty source file name", casFileName)
		}
	}
	if strings.ContainsAny(restoreName, `/\`) {
		return "", fmt.Errorf("restore from .cas failed: source file name %q contains a path", restoreName)
	}
	return restoreName, nil
}

func trimCASSuffix(name string) (string, bool) {
	const suffix = ".cas"
	if !strings.HasSuffix(strings.ToLower(name), suffix) {
		return "", false
	}
	return name[:len(name)-len(suffix)], true
}

func (y *Cloud189PC) restoreSourceFromCAS(ctx context.Context, dstDir model.Obj, file model.FileStreamer) (model.Obj, error) {
	info, err := readCASRestoreInfo(file)
	if err != nil {
		return nil, err
	}
	restoreName, err := y.resolveRestoreSourceName(file.GetName(), info)
	if err != nil {
		return nil, err
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
		"fileName":       url.QueryEscape(restoreName),
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
		return nil, fmt.Errorf("restore from .cas failed: source data for %q was not found in 189CloudPC", restoreName)
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
