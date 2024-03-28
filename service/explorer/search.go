package explorer

import (
	"context"
	"strings"

	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem"
	"github.com/Jaylenwa/Vfoy/pkg/hashid"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"
	"github.com/gin-gonic/gin"
)

// ItemSearchService 文件搜索服务
type ItemSearchService struct {
	Type     string `uri:"type" binding:"required"`
	Keywords string `uri:"keywords" binding:"required"`
	Path     string `form:"path"`
}

// Search 执行搜索
func (service *ItemSearchService) Search(c *gin.Context) serializer.Response {
	// 创建文件系统
	fs, err := filesystem.NewFileSystemFromContext(c)
	if err != nil {
		return serializer.Err(serializer.CodeCreateFSError, "", err)
	}
	defer fs.Recycle()

	if service.Path != "" {
		ok, parent := fs.IsPathExist(service.Path)
		if !ok {
			return serializer.Err(serializer.CodeParentNotExist, "", nil)
		}

		fs.Root = parent
	}

	switch service.Type {
	case "keywords":
		return service.SearchKeywords(c, fs, "%"+service.Keywords+"%")
	case "image":
		return service.SearchKeywords(c, fs, "%.bmp", "%.iff", "%.png", "%.gif", "%.jpg", "%.jpeg", "%.psd", "%.svg", "%.webp")
	case "video":
		return service.SearchKeywords(c, fs, "%.mp4", "%.flv", "%.avi", "%.wmv", "%.mkv", "%.rm", "%.rmvb", "%.mov", "%.ogv")
	case "audio":
		return service.SearchKeywords(c, fs, "%.mp3", "%.flac", "%.ape", "%.wav", "%.acc", "%.ogg", "%.midi", "%.mid")
	case "doc":
		return service.SearchKeywords(c, fs, "%.txt", "%.md", "%.pdf", "%.doc", "%.docx", "%.ppt", "%.pptx", "%.xls", "%.xlsx", "%.pub")
	case "tag":
		if tid, err := hashid.DecodeHashID(service.Keywords, hashid.TagID); err == nil {
			if tag, err := model.GetTagsByID(tid, fs.User.ID); err == nil {
				if tag.Type == model.FileTagType {
					exp := strings.Split(tag.Expression, "\n")
					expInput := make([]interface{}, len(exp))
					for i := 0; i < len(exp); i++ {
						expInput[i] = exp[i]
					}
					return service.SearchKeywords(c, fs, expInput...)
				}
			}
		}
		return serializer.Err(serializer.CodeNotFound, "", nil)
	default:
		return serializer.ParamErr("Unknown search type", nil)
	}
}

// SearchKeywords 根据关键字搜索文件
func (service *ItemSearchService) SearchKeywords(c *gin.Context, fs *filesystem.FileSystem, keywords ...interface{}) serializer.Response {
	// 上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取子项目
	objects, err := fs.Search(ctx, keywords...)
	if err != nil {
		return serializer.Err(serializer.CodeNotSet, err.Error(), err)
	}

	return serializer.Response{
		Code: 0,
		Data: map[string]interface{}{
			"parent":  0,
			"objects": objects,
		},
	}
}
