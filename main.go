// NovelAI项目主入口

package main

import (
	"log"
	
	"github.com/cloudwego/hertz/pkg/app/server"
	
	"novelai/biz/dal/db"
)

// 初始化PostgreSQL数据库
func initDB() {
	// 配置PostgreSQL连接
	dbConfig := &db.Config{
		DriverName: "postgres",
		// 根据实际情况修改DSN信息
		DSN:        "host=localhost user=postgres password=postgres dbname=novelai port=5432 sslmode=disable TimeZone=Asia/Shanghai",
		Active:     10, // 最大活跃连接数
		Idle:       5,  // 最大空闲连接数
	}
	
	// 初始化数据库连接
	db.Init(dbConfig)
	
	// 自动迁移数据库表结构
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("数据库表结构迁移失败: %v", err)
	}
}

func main() {
	// 初始化数据库
	initDB()
	log.Println("数据库初始化完成")
	
	// 创建Hertz服务器实例
	h := server.Default()
	
	// 注册路由
	register(h)
	
	// 启动服务器
	log.Println("开始启动API服务...")
	h.Spin()
}
