package middleware

import (
	"net/http"
	"strings"

	"github.com/Jaylenwa/Vfoy/pkg/cache"
	"github.com/Jaylenwa/Vfoy/pkg/sessionstore"

	"github.com/Jaylenwa/Vfoy/pkg/conf"
	"github.com/Jaylenwa/Vfoy/pkg/serializer"
	"github.com/Jaylenwa/Vfoy/pkg/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Store session存储
var Store sessions.Store

// Session 初始化session
func Session(secret string) gin.HandlerFunc {
	// Redis设置不为空，且非测试模式时使用Redis
	Store = sessionstore.NewStore(cache.Store, []byte(secret))

	sameSiteMode := http.SameSiteDefaultMode
	switch strings.ToLower(conf.CORSConfig.SameSite) {
	case "default":
		sameSiteMode = http.SameSiteDefaultMode
	case "none":
		sameSiteMode = http.SameSiteNoneMode
	case "strict":
		sameSiteMode = http.SameSiteStrictMode
	case "lax":
		sameSiteMode = http.SameSiteLaxMode
	}

	// Also set Secure: true if using SSL, you should though
	Store.Options(sessions.Options{
		HttpOnly: true,
		MaxAge:   60 * 86400,
		Path:     "/",
		SameSite: sameSiteMode,
		Secure:   conf.CORSConfig.Secure,
	})

	return sessions.Sessions("vfoy-session", Store)
}

// CSRFInit 初始化CSRF标记
func CSRFInit() gin.HandlerFunc {
	return func(c *gin.Context) {
		util.SetSession(c, map[string]interface{}{"CSRF": true})
		c.Next()
	}
}

// CSRFCheck 检查CSRF标记
func CSRFCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if check, ok := util.GetSession(c, "CSRF").(bool); ok && check {
			c.Next()
			return
		}

		c.JSON(200, serializer.Err(serializer.CodeNoPermissionErr, "Invalid origin", nil))
		c.Abort()
	}
}
