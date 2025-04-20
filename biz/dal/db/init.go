/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormopentracing "gorm.io/plugin/opentracing"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Config 数据库配置
type Config struct {
	DriverName string // 数据库驱动名称
	DSN        string // 数据源连接字符串
	Active     int    // 活跃连接数 
	Idle       int    // 空闲连接数
}

// Init 初始化数据库连接
// 该函数通过配置参数初始化数据库连接，支持不同数据库驱动
// 当前未选择具体数据库，保留驱动选择的灵活性
func Init(config *Config) {
	var err error
	var dialector gorm.Dialector

	// 根据驱动类型初始化对应的数据库连接
	// 支持未来扩展不同数据库类型
	switch config.DriverName {
	case "postgres":
		dialector = postgres.Open(config.DSN)
	// 可添加 MySQL、SQLite 等其他驱动支持
	// case "mysql":
	//     dialector = mysql.Open(config.DSN)
	// case "sqlite":
	//     dialector = sqlite.Open(config.DSN)
	default:
		log.Printf("未指定或不支持的数据库驱动类型: %s", config.DriverName)
		return
	}

	// 配置自定义日志记录器
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n[DB] ", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // 慢查询阈值
			LogLevel:                  logger.Info,   // 日志级别
			IgnoreRecordNotFoundError: true,          // 忽略记录未找到错误
			Colorful:                  true,          // 启用彩色打印
		},
	)

	// 初始化 GORM 数据库连接
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger:                  newLogger,      // 使用自定义日志记录器
		PrepareStmt:            true,           // 开启预编译语句缓存
		SkipDefaultTransaction: true,           // 跳过默认事务提高性能
	})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 添加 OpenTracing 插件支持，用于分布式追踪
	if err = DB.Use(gormopentracing.New()); err != nil {
		panic("OpenTracing 插件初始化失败: " + err.Error())
	}

	// 设置连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		panic("获取底层数据库连接失败: " + err.Error())
	}

	// 设置最大连接数
	if config.Active > 0 {
		sqlDB.SetMaxOpenConns(config.Active)
	}
	// 设置最大空闲连接数
	if config.Idle > 0 {
		sqlDB.SetMaxIdleConns(config.Idle)
	}

	log.Printf("数据库连接初始化成功")
}
