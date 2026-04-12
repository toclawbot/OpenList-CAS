package _189

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	Username                   string `json:"username" required:"true"`
	Password                   string `json:"password" required:"true"`
	Cookie                     string `json:"cookie" help:"Fill in the cookie if need captcha"`
	GenerateCAS                bool   `json:"generate_cas" help:"After upload, generate a same-name .cas file in the same directory"`
	DeleteSource               bool   `json:"delete_source" help:"After generating the .cas file, delete the uploaded source file"`
	RestoreSourceFromCAS       bool   `json:"restore_source_from_cas" help:"When uploading a .cas file, try to restore the source file by rapid upload instead of uploading the .cas file itself"`
	RestoreSourceUseCurrentName bool  `json:"restore_source_use_current_name" help:"When restoring from a .cas file, use the current .cas filename without the .cas suffix as the restored source filename"`
	GenerateCASAndDeleteSource bool   `json:"generate_cas_and_delete_source" ignore:"true"`
	driver.RootID
}

var config = driver.Config{
	Name:        "189Cloud",
	LocalSort:   true,
	DefaultRoot: "-11",
	Alert:       `info|You can try to use 189PC driver if this driver does not work.`,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Cloud189{}
	})
}
