package filesystem

import (
	"net/http/httptest"

	"github.com/Jaylenwa/Vfoy/pkg/cluster"
	"github.com/Jaylenwa/Vfoy/pkg/conf"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem/driver/shadow/masterinslave"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem/driver/shadow/slaveinmaster"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/Jaylenwa/Vfoy/models"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem/driver/local"
	"github.com/Jaylenwa/Vfoy/pkg/filesystem/driver/remote"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestNewFileSystem(t *testing.T) {
	asserts := assert.New(t)
	user := model.User{
		Policy: model.Policy{
			Type: "local",
		},
	}

	// 本地 成功
	fs, err := NewFileSystem(&user)
	asserts.NoError(err)
	asserts.NotNil(fs.Handler)
	asserts.IsType(local.Driver{}, fs.Handler)
	// 远程
	user.Policy.Type = "remote"
	fs, err = NewFileSystem(&user)
	asserts.NoError(err)
	asserts.NotNil(fs.Handler)
	asserts.IsType(&remote.Driver{}, fs.Handler)

	user.Policy.Type = "unknown"
	fs, err = NewFileSystem(&user)
	asserts.Error(err)
}

func TestNewFileSystemFromContext(t *testing.T) {
	asserts := assert.New(t)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user", &model.User{
		Policy: model.Policy{
			Type: "local",
		},
	})
	fs, err := NewFileSystemFromContext(c)
	asserts.NotNil(fs)
	asserts.NoError(err)

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	fs, err = NewFileSystemFromContext(c)
	asserts.Nil(fs)
	asserts.Error(err)
}

func TestDispatchHandler(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{
		User: &model.User{},
		Policy: &model.Policy{
			Type: "local",
		},
	}

	// 未指定，使用用户默认
	err := fs.DispatchHandler()
	asserts.NoError(err)
	asserts.IsType(local.Driver{}, fs.Handler)

	// 已指定，发生错误
	fs.Policy = &model.Policy{Type: "unknown"}
	err = fs.DispatchHandler()
	asserts.Error(err)

	fs.Policy = &model.Policy{Type: "mock"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "local"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "remote"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "qiniu"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "oss", Server: "https://s.com", BucketName: "1234"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "upyun"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "onedrive"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "cos"}
	err = fs.DispatchHandler()
	asserts.NoError(err)

	fs.Policy = &model.Policy{Type: "s3"}
	err = fs.DispatchHandler()
	asserts.NoError(err)
}

func TestNewFileSystemFromCallback(t *testing.T) {
	asserts := assert.New(t)

	// 用户上下文不存在
	{
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		fs, err := NewFileSystemFromCallback(c)
		asserts.Nil(fs)
		asserts.Error(err)
	}

	// 找不到回调会话
	{
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user", &model.User{
			Policy: model.Policy{
				Type: "local",
			},
		})
		fs, err := NewFileSystemFromCallback(c)
		asserts.Nil(fs)
		asserts.Error(err)
	}

	// 成功
	{
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user", &model.User{
			Policy: model.Policy{
				Type: "local",
			},
		})
		c.Set(UploadSessionCtx, &serializer.UploadSession{Policy: model.Policy{Type: "local"}})
		fs, err := NewFileSystemFromCallback(c)
		asserts.NotNil(fs)
		asserts.NoError(err)
	}

}

func TestFileSystem_SetTargetFileByIDs(t *testing.T) {
	asserts := assert.New(t)

	// 成功
	{
		fs := &FileSystem{}
		mock.ExpectQuery("SELECT(.+)").
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "1.txt"))
		err := fs.SetTargetFileByIDs([]uint{1, 2})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Len(fs.FileTarget, 1)
		asserts.NoError(err)
	}

	// 未找到
	{
		fs := &FileSystem{}
		mock.ExpectQuery("SELECT(.+)").WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		err := fs.SetTargetFileByIDs([]uint{1, 2})
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Len(fs.FileTarget, 0)
		asserts.Error(err)
	}
}

func TestFileSystem_CleanTargets(t *testing.T) {
	asserts := assert.New(t)
	fs := &FileSystem{
		FileTarget: []model.File{{}, {}},
		DirTarget:  []model.Folder{{}, {}},
	}

	fs.CleanTargets()
	asserts.Len(fs.FileTarget, 0)
	asserts.Len(fs.DirTarget, 0)
}

func TestNewAnonymousFileSystem(t *testing.T) {
	asserts := assert.New(t)

	// 正常
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policies"}).AddRow(3, "游客", "[]"))
		fs, err := NewAnonymousFileSystem()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal("游客", fs.User.Group.Name)
	}

	// 游客用户组不存在
	{
		mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "policies"}))
		fs, err := NewAnonymousFileSystem()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
		asserts.Nil(fs)
	}

	// 从机
	{
		conf.SystemConfig.Mode = "slave"
		fs, err := NewAnonymousFileSystem()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.NotNil(fs)
		asserts.NotNil(fs.Handler)
	}
}

func TestFileSystem_Recycle(t *testing.T) {
	fs := &FileSystem{
		User:       &model.User{},
		Policy:     &model.Policy{},
		FileTarget: []model.File{{}},
		DirTarget:  []model.Folder{{}},
		Hooks:      map[string][]Hook{"AfterUpload": {GenericAfterUpdate}},
	}
	fs.Recycle()
	newFS := getEmptyFS()
	if fs != newFS {
		t.Error("指针不一致")
	}
}

func TestFileSystem_SetTargetByInterface(t *testing.T) {
	asserts := assert.New(t)
	fs := FileSystem{}

	// 目录
	{
		asserts.NoError(fs.SetTargetByInterface(&model.Folder{}))
		asserts.Len(fs.DirTarget, 1)
		asserts.Len(fs.FileTarget, 0)
	}

	// 文件
	{
		asserts.NoError(fs.SetTargetByInterface(&model.File{}))
		asserts.Len(fs.DirTarget, 1)
		asserts.Len(fs.FileTarget, 1)
	}
}

func TestFileSystem_SwitchToSlaveHandler(t *testing.T) {
	a := assert.New(t)
	fs := FileSystem{
		User: &model.User{},
	}
	mockNode := &cluster.MasterNode{
		Model: &model.Node{},
	}
	fs.SwitchToSlaveHandler(mockNode)
	a.IsType(&slaveinmaster.Driver{}, fs.Handler)
}

func TestFileSystem_SwitchToShadowHandler(t *testing.T) {
	a := assert.New(t)
	fs := FileSystem{
		User:   &model.User{},
		Policy: &model.Policy{},
	}
	mockNode := &cluster.MasterNode{
		Model: &model.Node{},
	}

	// local to remote
	{
		fs.Policy.Type = "local"
		fs.SwitchToShadowHandler(mockNode, "", "")
		a.IsType(&masterinslave.Driver{}, fs.Handler)
	}

	// onedrive
	{
		fs.Policy.Type = "onedrive"
		fs.SwitchToShadowHandler(mockNode, "", "")
		a.IsType(&masterinslave.Driver{}, fs.Handler)
	}
}
