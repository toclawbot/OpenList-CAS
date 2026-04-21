package openlistplus

import (
	"io"

	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/openlistplus/casfile"
)

func ParseCASBytes(data []byte) (*casfile.Info, error) {
	return casfile.Parse(data)
}

func ParseCASFile(file model.FileStreamer) (*casfile.Info, error) {
	cache, err := file.CacheFullAndWriter(nil, nil)
	if err != nil {
		return nil, err
	}
	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(cache)
	if err != nil {
		return nil, err
	}
	if _, err = cache.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return casfile.Parse(data)
}
