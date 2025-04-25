/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"log"
)

// AutoMigrate 自动迁移数据库表结构
// 该函数会根据模型定义自动创建或更新数据库表
// 只在应用启动时调用一次
func AutoMigrate() error {
	log.Println("开始自动迁移数据库表结构...")

	// 添加需要迁移的模型（如有新表需在此处追加）
	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Printf("迁移用户表失败: %v", err)
		return err
	}
	if err := DB.AutoMigrate(&Save{}); err != nil {
		log.Printf("迁移保存表失败: %v", err)
		return err
	}

	log.Println("数据库表结构迁移完成")
	return nil
}
