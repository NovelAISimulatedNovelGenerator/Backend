// 自动生成的路由文件，请根据需要修改

package save

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	handler "novelai/biz/handler/save"
)

// 注册保存相关路由
func RegisterRoutes(r *server.Hertz) {
	saveGroup := r.Group("/api/save")
	{
		saveGroup.POST("/create", handler.CreateSave)
		saveGroup.GET("/get", handler.GetSave)
		saveGroup.PUT("/update", handler.UpdateSave)
		saveGroup.DELETE("/delete", handler.DeleteSave)
		saveGroup.GET("/list", handler.ListSaves)
	}
}
