package _189pc

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	LoginType    string `json:"login_type" type:"select" options:"password,qrcode" default:"password" required:"true"`
	Username     string `json:"username" required:"true"`
	Password     string `json:"password" required:"true"`
	VCode        string `json:"validate_code"`
	RefreshToken string `json:"refresh_token" help:"To switch accounts, please clear this field"`
	driver.RootID
	OrderBy                    string `json:"order_by" type:"select" options:"filename,filesize,lastOpTime" default:"filename"`
	OrderDirection             string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
	Type                       string `json:"type" type:"select" options:"personal,family" default:"personal"`
	FamilyID                   string `json:"family_id"`
	UploadMethod               string `json:"upload_method" type:"select" options:"stream,rapid,old" default:"stream"`
	UploadThread               string `json:"upload_thread" default:"3" help:"1<=thread<=32"`
	FamilyTransfer             bool   `json:"family_transfer"`
	RapidUpload                bool   `json:"rapid_upload"`
	NoUseOcr                   bool   `json:"no_use_ocr"`
	GenerateCAS                bool   `json:"generate_cas" help:"After upload, generate a same-name .cas file in the same directory"`
	DeleteSource               bool   `json:"delete_source" help:"After generating the .cas file, delete the uploaded source file"`
	RestoreSourceFromCAS       bool   `json:"restore_source_from_cas" help:"When uploading a .cas file, try to restore the source file by rapid upload instead of uploading the .cas file itself"`
	GenerateCASAndDeleteSource bool   `json:"generate_cas_and_delete_source" ignore:"true"`
}

var config = driver.Config{
	Name:        "189CloudPC",
	DefaultRoot: "-11",
	CheckStatus: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Cloud189PC{}
	})
}
