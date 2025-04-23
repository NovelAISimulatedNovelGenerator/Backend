// 自动生成的路由文件，请根据需要修改

package save

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	handler "novelai/biz/handler/save"
)

// 注册保存相关路由
import "novelai/pkg/middleware"

func RegisterRoutes(r *server.Hertz) {
	jwtMw, err := middleware.JwtMiddleware()
	if err != nil {
		panic("JWT中间件初始化失败: " + err.Error())
	}
	saveGroup := r.Group("/api/save")
	saveGroup.Use(jwtMw.MiddlewareFunc())
	{
		saveGroup.POST("/create", handler.CreateSave)
		saveGroup.GET("/get", handler.GetSave)
		saveGroup.PUT("/update", handler.UpdateSave)
		saveGroup.DELETE("/delete", handler.DeleteSave)
		saveGroup.GET("/list", handler.ListSaves)
	}
}
