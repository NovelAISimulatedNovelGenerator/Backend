// 自动生成的路由文件，请根据需要修改

package user

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	handler "novelai/biz/handler/user"
)

// 注册用户相关路由
func RegisterRoutes(r *server.Hertz) {
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
		userGroup.GET("/info", handler.GetUser)
		userGroup.PUT("/update", handler.UpdateUser)
	}
}
