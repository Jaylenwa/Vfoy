package scripts

import (
	"context"

	model "github.com/Jaylenwa/Vfoy/v3/models"
	"github.com/Jaylenwa/Vfoy/v3/pkg/util"
	"github.com/fatih/color"
)

type ResetAdminPassword int

// Run 运行脚本从社区版升级至 Pro 版
func (script ResetAdminPassword) Run(ctx context.Context) {
	// 查找用户
	user, err := model.GetUserByID(1)
	if err != nil {
		util.Log().Panic("Initial admin user not exist: %s", err)
	}

	// 生成密码
	password := util.RandStringRunes(8)

	// 更改为新密码
	user.SetPassword(password)
	if err := user.Update(map[string]interface{}{"password": user.Password}); err != nil {
		util.Log().Panic("Failed to update password: %s", err)
	}

	c := color.New(color.FgWhite).Add(color.BgBlack).Add(color.Bold)
	util.Log().Info("Initial admin user password changed to:" + c.Sprint(password))
}
