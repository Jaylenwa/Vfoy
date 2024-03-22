package explorer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	model "github.com/Jaylenwa/Vfoy/v3/models"
	"github.com/Jaylenwa/Vfoy/v3/pkg/cache"
	"github.com/Jaylenwa/Vfoy/v3/pkg/cluster"
	"github.com/Jaylenwa/Vfoy/v3/pkg/filesystem"
	"github.com/Jaylenwa/Vfoy/v3/pkg/filesystem/fsctx"
	"github.com/Jaylenwa/Vfoy/v3/pkg/serializer"
	"github.com/Jaylenwa/Vfoy/v3/pkg/task"
	"github.com/Jaylenwa/Vfoy/v3/pkg/task/slavetask"
	"github.com/Jaylenwa/Vfoy/v3/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SlaveDownloadService 从机文件下載服务
type SlaveDownloadService struct {
	PathEncoded string `uri:"path" binding:"required"`
	Name        string `uri:"name" binding:"required"`
	Speed       int    `uri:"speed" binding:"min=0"`
}

// SlaveFileService 从机单文件文件相关服务
type SlaveFileService struct {
	PathEncoded string `uri:"path" binding:"required"`
	Ext         string `uri:"ext"`
}

// SlaveFilesService 从机多文件相关服务
type SlaveFilesService struct {
	Files []string `json:"files" binding:"required,gt=0"`
}

// SlaveListService 从机列表服务
type SlaveListService struct {
	Path      string `json:"path" binding:"required,min=1,max=65535"`
	Recursive bool   `json:"recursive"`
}

// ServeFile 通过签名的URL下载从机文件
func (service *SlaveDownloadService) ServeFile(ctx context.Context, c *gin.Context, isDownload bool) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 解码文件路径
	fileSource, err := base64.RawURLEncoding.DecodeString(service.PathEncoded)
	if err != nil {
		return serializer.Err(serializer.CodeFileNotFound, "", err)
	}

	// 根据URL里的信息创建一个文件对象和用户对象
	file := model.File{
		Name:       service.Name,
		SourceName: string(fileSource),
		Policy: model.Policy{
			Model: gorm.Model{ID: 1},
			Type:  "local",
		},
	}
	fs.User = &model.User{
		Group: model.Group{SpeedLimit: service.Speed},
	}
	fs.FileTarget = []model.File{file}

	// 开始处理下载
	ctx = context.WithValue(ctx, fsctx.GinCtx, c)
	rs, err := fs.GetDownloadContent(ctx, 0)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}
	defer rs.Close()

	// 设置下载文件名
	if isDownload {
		c.Header("Content-Disposition", "attachment; filename=\""+url.PathEscape(fs.FileTarget[0].Name)+"\"")
	}

	// 发送文件
	http.ServeContent(c.Writer, c.Request, fs.FileTarget[0].Name, time.Now(), rs)

	return serializer.Response{}
}

// Delete 通过签名的URL删除从机文件
func (service *SlaveFilesService) Delete(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 删除文件
	failed, err := fs.Handler.Delete(ctx, service.Files)
	if err != nil {
		// 将Data字段写为字符串方便主控端解析
		data, _ := json.Marshal(serializer.RemoteDeleteRequest{Files: failed})

		return serializer.Response{
			Code:  serializer.CodeNotFullySuccess,
			Data:  string(data),
			Msg:   fmt.Sprintf("Failed to delete %d files(s)", len(failed)),
			Error: err.Error(),
		}
	}
	return serializer.Response{}
}

// Thumb 通过签名URL获取从机文件缩略图
func (service *SlaveFileService) Thumb(ctx context.Context, c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewAnonymousFileSystem()
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 解码文件路径
	fileSource, err := base64.RawURLEncoding.DecodeString(service.PathEncoded)
	if err != nil {
		return serializer.Err(serializer.CodeFileNotFound, "", err)
	}
	fs.FileTarget = []model.File{{SourceName: string(fileSource), Name: fmt.Sprintf("%s.%s", fileSource, service.Ext), PicInfo: "1,1"}}

	// 获取缩略图
	resp, err := fs.GetThumb(ctx, 0)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, "Failed to get thumb", err)
	}

	defer resp.Content.Close()
	http.ServeContent(c.Writer, c.Request, "thumb.png", time.Now(), resp.Content)

	return serializer.Response{}
}

// CreateTransferTask 创建从机文件转存任务
func CreateTransferTask(c *gin.Context, req *serializer.SlaveTransferReq) serializer.Response {
	if id, ok := c.Get("MasterSiteID"); ok {
		job := &slavetask.TransferTask{
			Req:      req,
			MasterID: id.(string),
		}

		if err := cluster.DefaultController.SubmitTask(job.MasterID, job, req.Hash(job.MasterID), func(job interface{}) {
			task.TaskPoll.Submit(job.(task.Job))
		}); err != nil {
			return serializer.Err(serializer.CodeCreateTaskError, "", err)
		}

		return serializer.Response{}
	}

	return serializer.ParamErr("未知的主机节点ID", nil)
}

// SlaveListService 从机上传会话服务
type SlaveCreateUploadSessionService struct {
	Session   serializer.UploadSession `json:"session" binding:"required"`
	TTL       int64                    `json:"ttl"`
	Overwrite bool                     `json:"overwrite"`
}

// Create 从机创建上传会话
func (service *SlaveCreateUploadSessionService) Create(ctx context.Context, c *gin.Context) serializer.Response {
	if !service.Overwrite && util.Exists(service.Session.SavePath) {
		return serializer.Err(serializer.CodeConflict, "placeholder file already exist", nil)
	}

	err := cache.Set(
		filesystem.UploadSessionCachePrefix+service.Session.Key,
		service.Session,
		int(service.TTL),
	)
	if err != nil {
		return serializer.Err(serializer.CodeCacheOperation, "Failed to create upload session in slave node", err)
	}

	return serializer.Response{}
}
