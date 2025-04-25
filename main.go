// NovelAI项目主入口

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"novelai/biz/dal/db"
)

// 获取环境变量值，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// 初始化PostgreSQL数据库
func initDB() {
	// 从环境变量获取数据库连接信息
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "novelai")
	dbTimeZone := getEnv("DB_TIMEZONE", "Asia/Shanghai")
	
	// 构建DSN连接字符串
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbTimeZone,
	)
	
	// 配置PostgreSQL连接
	dbConfig := &db.Config{
		DriverName: "postgres",
		DSN:        dsn,
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
	// 设置 Hertz 日志级别为 debug，便于调试和问题追踪
	hlog.SetLevel(hlog.LevelDebug)

	// 初始化数据库
	initDB()
	hlog.Debug("数据库初始化完成")

	// 创建Hertz服务器实例
	h := server.Default()
	hlog.Debug("Hertz 服务器实例创建完成")

	// 注册路由
	register(h)
	hlog.Debug("路由注册完成")

	// 启动服务器
	hlog.Debug("开始启动API服务...")
	h.Spin()
}
