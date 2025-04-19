/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package dal

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"novelai/biz/dal/db"
)

// Config 数据访问层配置结构
type Config struct {
	DB *db.Config // 数据库配置
}

// Init 初始化数据访问层
// 该函数负责初始化所有数据存储相关组件，包括数据库连接
// 参数：
//   - config: 数据访问层配置
func Init(config *Config) {
	// 检查配置是否有效
	if config == nil {
		hlog.Warn("DAL 配置为空，将使用默认配置")
		config = &Config{
			DB: &db.Config{},
		}
	}

	// 初始化数据库连接
	if config.DB != nil {
		db.Init(config.DB)
	} else {
		hlog.Warn("数据库配置为空，数据库连接未初始化")
	}

	// 这里可以添加其他存储服务的初始化，如 Redis、MongoDB 等
	hlog.Info("数据访问层初始化完成")
}
