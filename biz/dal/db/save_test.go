/*
 * NovelAI Project
 * Copyright (C) 2023-2025
 */

package db

import (
	"fmt"
	"testing"
	"time"

	"novelai/pkg/constants"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 测试初始化函数，使用SQLite内存数据库
func setupSaveTestDB(t *testing.T) {
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "初始化测试数据库失败")

	err = DB.AutoMigrate(&Save{})
	assert.NoError(t, err, "自动迁移存档表失败")

	DB.Exec("DELETE FROM " + constants.TableNameSave)
}

// 创建测试存档
func createTestSave(t *testing.T, userID int64) *Save {
	timestamp := time.Now().UnixNano()
	saveID := fmt.Sprintf("saveid-%d-%d", userID, timestamp)
	saveName := fmt.Sprintf("测试存档%d-%d", userID, timestamp)
	save := &Save{
		UserID:          userID,
		SaveID:          saveID,
		SaveName:        saveName,
		SaveDescription: "这是测试存档描述",
		SaveData:        "{\"key\":\"value\"}",
		SaveType:        "draft",
		SaveStatus:      "active",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
	}
	id, err := CreateSave(save)
	assert.NoError(t, err, "创建测试存档失败")
	assert.Greater(t, id, int64(0), "存档ID应大于0")

	createdSave, err := QuerySaveByID(id)
	assert.NoError(t, err, "查询创建的存档失败")
	return createdSave
}

// TestCreateSave 测试存档创建
func TestCreateSave(t *testing.T) {
	setupSaveTestDB(t)
	timestamp := time.Now().UnixNano()
	save := &Save{
		UserID:          1,
		SaveID:          fmt.Sprintf("saveid-%d", timestamp),
		SaveName:        fmt.Sprintf("存档1-%d", timestamp),
		SaveDescription: "描述1",
		SaveData:        "{}",
		SaveType:        "draft",
		SaveStatus:      "active",
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
	}
	id, err := CreateSave(save)
	assert.NoError(t, err, "创建存档失败")
	assert.Greater(t, id, int64(0), "存档ID应大于0")
}

// TestQuerySaveByID 测试通过ID查询存档
func TestQuerySaveByID(t *testing.T) {
	setupSaveTestDB(t)
	save := createTestSave(t, 1)
	queried, err := QuerySaveByID(save.ID)
	assert.NoError(t, err, "通过ID查询存档失败")
	assert.Equal(t, save.SaveName, queried.SaveName)
}

// TestQuerySavesByUser 测试按用户ID分页查询
func TestQuerySavesByUser(t *testing.T) {
	setupSaveTestDB(t)
	userID := int64(2)
	for i := 0; i < 5; i++ {
		createTestSave(t, userID)
	}
	saves, total, err := QuerySavesByUser(userID, 1, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, saves, 3)
}

// TestUpdateSave 测试更新存档
func TestUpdateSave(t *testing.T) {
	setupSaveTestDB(t)
	save := createTestSave(t, 3)
	save.SaveName = "更新后的名称"
	save.SaveDescription = "更新后的描述"
	save.SaveData = "{\"key\":\"new\"}"
	save.SaveType = "config"
	save.SaveStatus = "deleted"
	err := UpdateSave(save)
	assert.NoError(t, err)
	updated, err := QuerySaveByID(save.ID)
	assert.NoError(t, err)
	assert.Equal(t, "更新后的名称", updated.SaveName)
	assert.Equal(t, "更新后的描述", updated.SaveDescription)
	assert.Equal(t, "{\"key\":\"new\"}", updated.SaveData)
	assert.Equal(t, "config", updated.SaveType)
	assert.Equal(t, "deleted", updated.SaveStatus)
}

// TestDeleteSave 测试删除存档
func TestDeleteSave(t *testing.T) {
	setupSaveTestDB(t)
	save := createTestSave(t, 4)
	err := DeleteSave(save.ID)
	assert.NoError(t, err)
	_, err = QuerySaveByID(save.ID)
	assert.ErrorIs(t, err, ErrSaveNotFound)
}

// TestListSaves 测试分页获取所有存档
func TestListSaves(t *testing.T) {
	setupSaveTestDB(t)
	for i := 0; i < 7; i++ {
		createTestSave(t, 5)
	}
	saves, total, err := ListSaves(1, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), total)
	assert.Len(t, saves, 5)
}

// TestCheckSaveExists 测试检查存档是否存在
func TestCheckSaveExists(t *testing.T) {
	setupSaveTestDB(t)
	save := createTestSave(t, 6)
	exists, err := CheckSaveExists(save.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	notExists, err := CheckSaveExists(99999)
	assert.NoError(t, err)
	assert.False(t, notExists)
}
