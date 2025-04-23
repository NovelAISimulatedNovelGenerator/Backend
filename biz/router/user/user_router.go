// 自动生成的路由文件，请根据需要修改

package user

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	handler "novelai/biz/handler/user"
)

// 注册用户相关路由
import "novelai/pkg/middleware"

func RegisterRoutes(r *server.Hertz) {
	jwtMw, err := middleware.JwtMiddleware()
	if err != nil {
		panic("JWT中间件初始化失败: " + err.Error())
	}
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", jwtMw.LoginHandler)
		userGroup.Use(jwtMw.MiddlewareFunc())
		userGroup.GET("/info", handler.GetUser)
		userGroup.PUT("/update", handler.UpdateUser)
	}
}
