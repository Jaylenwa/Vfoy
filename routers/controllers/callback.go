package controllers

import (
	"path"
	"strconv"

	model "github.com/Jaylenwa/Vfoy/v3/models"

	"github.com/Jaylenwa/Vfoy/v3/pkg/serializer"
	"github.com/Jaylenwa/Vfoy/v3/pkg/util"
	"github.com/Jaylenwa/Vfoy/v3/service/callback"
	"github.com/gin-gonic/gin"
)

// RemoteCallback 远程上传回调
func RemoteCallback(c *gin.Context) {
	var callbackBody callback.RemoteUploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callback.ProcessCallback(callbackBody, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// QiniuCallback 七牛上传回调
func QiniuCallback(c *gin.Context) {
	var callbackBody callback.UploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callback.ProcessCallback(callbackBody, c)
		if res.Code != 0 {
			c.JSON(401, serializer.GeneralUploadCallbackFailed{Error: res.Msg})
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(401, ErrorResponse(err))
	}
}

// OSSCallback 阿里云OSS上传回调
func OSSCallback(c *gin.Context) {
	var callbackBody callback.UploadCallbackService
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		if callbackBody.PicInfo == "," {
			callbackBody.PicInfo = ""
		}
		res := callback.ProcessCallback(callbackBody, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UpyunCallback 又拍云上传回调
func UpyunCallback(c *gin.Context) {
	var callbackBody callback.UpyunCallbackService
	if err := c.ShouldBind(&callbackBody); err == nil {
		if callbackBody.Code != 200 {
			util.Log().Debug(
				"Upload callback returned error code:%d, message: %s",
				callbackBody.Code,
				callbackBody.Message,
			)
			return
		}
		res := callback.ProcessCallback(callbackBody, c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// OneDriveCallback OneDrive上传完成客户端回调
func OneDriveCallback(c *gin.Context) {
	var callbackBody callback.OneDriveCallback
	if err := c.ShouldBindJSON(&callbackBody); err == nil {
		res := callbackBody.PreProcess(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// OneDriveOAuth OneDrive 授权回调
func OneDriveOAuth(c *gin.Context) {
	var callbackBody callback.OauthService
	if err := c.ShouldBindQuery(&callbackBody); err == nil {
		res := callbackBody.OdAuth(c)
		redirect := model.GetSiteURL()
		redirect.Path = path.Join(redirect.Path, "/admin/policy")
		queries := redirect.Query()
		queries.Add("code", strconv.Itoa(res.Code))
		queries.Add("msg", res.Msg)
		queries.Add("err", res.Error)
		redirect.RawQuery = queries.Encode()
		c.Redirect(303, redirect.String())
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// GoogleDriveOAuth Google Drive 授权回调
func GoogleDriveOAuth(c *gin.Context) {
	var callbackBody callback.OauthService
	if err := c.ShouldBindQuery(&callbackBody); err == nil {
		res := callbackBody.GDriveAuth(c)
		redirect := model.GetSiteURL()
		redirect.Path = path.Join(redirect.Path, "/admin/policy")
		queries := redirect.Query()
		queries.Add("code", strconv.Itoa(res.Code))
		queries.Add("msg", res.Msg)
		queries.Add("err", res.Error)
		redirect.RawQuery = queries.Encode()
		c.Redirect(303, redirect.String())
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// COSCallback COS上传完成客户端回调
func COSCallback(c *gin.Context) {
	var callbackBody callback.COSCallback
	if err := c.ShouldBindQuery(&callbackBody); err == nil {
		res := callbackBody.PreProcess(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// S3Callback S3上传完成客户端回调
func S3Callback(c *gin.Context) {
	var callbackBody callback.S3Callback
	if err := c.ShouldBindQuery(&callbackBody); err == nil {
		res := callbackBody.PreProcess(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
