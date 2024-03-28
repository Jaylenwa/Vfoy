package bootstrap

import (
	"io/fs"
	"path/filepath"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/models/scripts"
	"github.com/Jaylenwa/Vfoy/pkg/aria2"
	"github.com/Jaylenwa/Vfoy/pkg/auth"
	"github.com/Jaylenwa/Vfoy/pkg/cache"
	"github.com/Jaylenwa/Vfoy/pkg/cluster"
	"github.com/Jaylenwa/Vfoy/pkg/conf"
	"github.com/Jaylenwa/Vfoy/pkg/crontab"
	"github.com/Jaylenwa/Vfoy/pkg/email"
	"github.com/Jaylenwa/Vfoy/pkg/mq"
	"github.com/Jaylenwa/Vfoy/pkg/task"
	"github.com/Jaylenwa/Vfoy/pkg/wopi"
	"github.com/gin-gonic/gin"
)

// Init 初始化启动
func Init(path string, statics fs.FS) {
	InitApplication()
	conf.Init(path)
	// Debug 关闭时，切换为生产模式
	if !conf.SystemConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	dependencies := []struct {
		mode    string
		factory func()
	}{
		{
			"both",
			func() {
				scripts.Init()
			},
		},
		{
			"both",
			func() {
				cache.Init()
			},
		},
		{
			"slave",
			func() {
				model.InitSlaveDefaults()
			},
		},
		{
			"slave",
			func() {
				cache.InitSlaveOverwrites()
			},
		},
		{
			"master",
			func() {
				model.Init()
			},
		},
		{
			"both",
			func() {
				cache.Restore(filepath.Join(model.GetSettingByName("temp_path"), cache.DefaultCacheFile))
			},
		},
		{
			"both",
			func() {
				task.Init()
			},
		},
		{
			"master",
			func() {
				cluster.Init()
			},
		},
		{
			"master",
			func() {
				aria2.Init(false, cluster.Default, mq.GlobalMQ)
			},
		},
		{
			"master",
			func() {
				email.Init()
			},
		},
		{
			"master",
			func() {
				crontab.Init()
			},
		},
		{
			"master",
			func() {
				InitStatic(statics)
			},
		},
		{
			"slave",
			func() {
				cluster.InitController()
			},
		},
		{
			"both",
			func() {
				auth.Init()
			},
		},
		{
			"master",
			func() {
				wopi.Init()
			},
		},
	}

	for _, dependency := range dependencies {
		if dependency.mode == conf.SystemConfig.Mode || dependency.mode == "both" {
			dependency.factory()
		}
	}

}
