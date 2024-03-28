package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Jaylenwa/Vfoy/bootstrap"
	"github.com/Jaylenwa/Vfoy/pkg/cache"
	"github.com/Jaylenwa/Vfoy/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	testMock "github.com/stretchr/testify/mock"
)

type StaticMock struct {
	testMock.Mock
}

func (m StaticMock) Open(name string) (http.File, error) {
	args := m.Called(name)
	return args.Get(0).(http.File), args.Error(1)
}

func (m StaticMock) Exists(prefix string, filepath string) bool {
	args := m.Called(prefix, filepath)
	return args.Bool(0)
}

func TestFrontendFileHandler(t *testing.T) {
	asserts := assert.New(t)
	rec := httptest.NewRecorder()

	// 静态资源未加载
	{
		TestFunc := FrontendFileHandler()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// index.html 不存在
	{
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(&os.File{}, errors.New("error"))
		TestFunc := FrontendFileHandler()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// index.html 读取失败
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		TestFunc := FrontendFileHandler()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)
		TestFunc(c)
		asserts.False(c.IsAborted())
	}

	// 成功且命中
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		defer file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		TestFunc := FrontendFileHandler()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/", nil)

		cache.Set("setting_siteName", "vfoy", 0)
		cache.Set("setting_siteKeywords", "vfoy", 0)
		cache.Set("setting_siteScript", "vfoy", 0)
		cache.Set("setting_pwa_small_icon", "vfoy", 0)

		TestFunc(c)
		asserts.True(c.IsAborted())
	}

	// 成功且命中静态文件
	{
		file, _ := util.CreatNestedFile("tests/index.html")
		defer file.Close()
		testStatic := &StaticMock{}
		bootstrap.StaticFS = testStatic
		testStatic.On("Open", "/index.html").
			Return(file, nil)
		testStatic.On("Exists", "/", "/2").
			Return(true)
		testStatic.On("Open", "/2").
			Return(file, nil)
		TestFunc := FrontendFileHandler()

		c, _ := gin.CreateTestContext(rec)
		c.Params = []gin.Param{}
		c.Request, _ = http.NewRequest("GET", "/2", nil)

		TestFunc(c)
		asserts.True(c.IsAborted())
		testStatic.AssertExpectations(t)
	}

	// API 相关跳过
	{
		for _, reqPath := range []string{"/api/user", "/manifest.json", "/dav/path"} {
			file, _ := util.CreatNestedFile("tests/index.html")
			defer file.Close()
			testStatic := &StaticMock{}
			bootstrap.StaticFS = testStatic
			testStatic.On("Open", "/index.html").
				Return(file, nil)
			TestFunc := FrontendFileHandler()

			c, _ := gin.CreateTestContext(rec)
			c.Params = []gin.Param{}
			c.Request, _ = http.NewRequest("GET", reqPath, nil)

			TestFunc(c)
			asserts.False(c.IsAborted())
		}
	}

}
