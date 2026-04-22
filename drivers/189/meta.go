package _189

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
	Username                    string `json:"username" required:"true"`
	Password                    string `json:"password" required:"true"`
	Cookie                      string `json:"cookie" help:"如需验证码时可填写 Cookie"`
	GenerateCAS                bool   `json:"generate_cas" help:"上传文件后，在同目录生成同名的 .cas 文件"`
	DeleteSource               bool   `json:"delete_source" help:"成功生成 .cas 文件后，自动删除源文件"`
	RestoreSourceFromCAS       bool   `json:"restore_source_from_cas" help:"上传 .cas 文件时，尝试根据其中的哈希信息秒传还原源文件"`
	RestoreSourceUseCurrentName bool  `json:"restore_source_use_current_name" help:"从 .cas 还原源文件时，使用当前 .cas 文件名（去除 .cas 后缀）；若无扩展名，则补充原始扩展名"`
	DeleteCASAfterRestore      bool   `json:"delete_cas_after_restore" help:"成功还原源文件后，自动删除 .cas 文件"`
	AutoRestoreExistingCAS     bool   `json:"auto_restore_existing_cas" help:"自动监测已配置目录中的 .cas 文件，检测到 .cas 文件时还原源文件"`
	AutoRestoreExistingCASPaths string `json:"auto_restore_existing_cas_paths" type:"text" help:"要监测的目录路径，每行一个，路径相对于当前存储根目录；会自动包含其下所有子目录"`
	GenerateCASAndDeleteSource  bool   `json:"generate_cas_and_delete_source" ignore:"true"`
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
