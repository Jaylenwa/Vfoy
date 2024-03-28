package explorer

import (
	"context"

	"github.com/Jaylenwa/Vfoy/pkg/filesystem"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// DirectoryService 创建新目录服务
type DirectoryService struct {
	Path string `uri:"path" json:"path" binding:"required,min=1,max=65535"`
}

// ListDirectory 列出目录内容
func (service *DirectoryService) ListDirectory(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取子项目
	objects, err := fs.List(ctx, service.Path, nil)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	var parentID uint
	if len(fs.DirTarget) > 0 {
		parentID = fs.DirTarget[0].ID
	}

	return serializer.Response{
		Code: 0,
		Data: serializer.BuildObjectList(parentID, objects, fs.Policy),
	}
}

// CreateDirectory 创建目录
func (service *DirectoryService) CreateDirectory(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建目录
	_, err = fs.CreateDirectory(ctx, service.Path)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFolderFailed, err.Error(), err)
	}
	return serializer.Response{
		Code: 0,
	}

}
