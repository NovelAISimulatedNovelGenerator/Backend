package db

import (
	"context"
	// "strconv" // Not used in the abridged version
	"testing"
	"time"

	"novelai/pkg/constants"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDBForWorldview initializes the database for Worldview tests.
func setupTestDBForWorldview(t *testing.T) {
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database for Worldview")

	err = DB.AutoMigrate(&Worldview{})
	assert.NoError(t, err, "Failed to migrate Worldview table")

	result := DB.Exec("DELETE FROM " + constants.TableNameWorldview)
	assert.NoError(t, result.Error, "Failed to clean Worldview table")
}

// createTestWorldview creates a test worldview instance and persists it.
func createTestWorldview(t *testing.T, nameSuffix string, parentID int64, tag string) *Worldview {
	ctx := context.Background()
	wv := &Worldview{
		Name:        "Test Worldview " + nameSuffix,
		Description: "Desc for " + nameSuffix,
		Tag:         tag,
		ParentID:    parentID,
	}
	id, err := CreateWorldview(ctx, wv)
	assert.NoError(t, err, "CreateWorldview failed for "+nameSuffix)
	assert.Greater(t, id, int64(0))

	createdWv, err := GetWorldviewByID(ctx, id)
	assert.NoError(t, err, "GetWorldviewByID failed after create for "+nameSuffix)
	assert.NotNil(t, createdWv)
	return createdWv
}

func TestCRUDWorldview(t *testing.T) {
	setupTestDBForWorldview(t)
	ctx := context.Background()

	// Test Create
	wv1 := &Worldview{
		Name:        "Primary Worldview",
		Description: "This is the main worldview.",
		Tag:         "primary,core",
		ParentID:    0,
	}
	id1, err := CreateWorldview(ctx, wv1)
	assert.NoError(t, err)
	assert.Greater(t, id1, int64(0))

	fetchedWv1, err := GetWorldviewByID(ctx, id1)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedWv1)
	assert.Equal(t, wv1.Name, fetchedWv1.Name)

	// Test Get Non-Existing
	_, err = GetWorldviewByID(ctx, 99999)
	assert.ErrorIs(t, err, ErrWorldviewNotFound)

	// Test Update
	updates := map[string]interface{}{
		"name":        "Updated Worldview Name",
		"description": "Updated worldview desc.",
		"tag":         "updated,new_tag",
	}
	oldUpdatedAt := fetchedWv1.UpdatedAt
	time.Sleep(1100 * time.Millisecond) // Ensure UpdatedAt changes (second precision for int64 timestamp)
	err = UpdateWorldview(ctx, id1, updates)
	assert.NoError(t, err)

	updatedWv, err := GetWorldviewByID(ctx, id1)
	assert.NoError(t, err)
	assert.Equal(t, updates["name"], updatedWv.Name)
	assert.Equal(t, updates["tag"], updatedWv.Tag)
	assert.Greater(t, updatedWv.UpdatedAt, oldUpdatedAt, "UpdatedAt timestamp (%v) should be greater than oldUpdatedAt (%v) after update", updatedWv.UpdatedAt, oldUpdatedAt)

	// Test Update Non-Existing
	err = UpdateWorldview(ctx, 99999, updates)
	assert.Error(t, err)

	// Test Delete
	err = DeleteWorldview(ctx, id1)
	assert.NoError(t, err)
	_, err = GetWorldviewByID(ctx, id1)
	assert.ErrorIs(t, err, ErrWorldviewNotFound)

	// Test Delete Non-Existing
	err = DeleteWorldview(ctx, 99999)
	assert.ErrorIs(t, err, ErrWorldviewNotFound)
}

func TestListWorldviews(t *testing.T) {
	setupTestDBForWorldview(t)
	ctx := context.Background()

	wvParent := createTestWorldview(t, "ParentList", 0, "parent,main")
	_ = createTestWorldview(t, "Child1P", wvParent.ID, "child,tagA")
	_ = createTestWorldview(t, "Child2P", wvParent.ID, "child,tagB")
	_ = createTestWorldview(t, "Top1", 0, "top,tagA")
	_ = createTestWorldview(t, "Top2", 0, "top,tagC")

	// Case 1: List children of wvParent (parentIDFilter = wvParent.ID)
	list1, total1, err1 := ListWorldviews(ctx, wvParent.ID, "", 1, 5)
	assert.NoError(t, err1)
	assert.Equal(t, int64(2), total1)
	assert.Len(t, list1, 2)

	// Case 2: List top-level worldviews (parentIDFilter = 0) - wvParent, Top1, Top2
	list2, total2, err2 := ListWorldviews(ctx, 0, "", 1, 5)
	assert.NoError(t, err2)
	assert.Equal(t, int64(3), total2)
	assert.Len(t, list2, 3)

	// Case 3: Filter by tag "tagA" (no parent filter: parentIDFilter = -1)
	list3, total3, err3 := ListWorldviews(ctx, -1, "tagA", 1, 5)
	assert.NoError(t, err3)
	assert.Equal(t, int64(2), total3, "Should find 2 worldviews with 'tagA'")
	assert.Len(t, list3, 2)

	// Case 4: Filter by tag "child" and parentIDFilter = wvParent.ID
	list4, total4, err4 := ListWorldviews(ctx, wvParent.ID, "child", 1, 5)
	assert.NoError(t, err4)
	assert.Equal(t, int64(2), total4)
	assert.Len(t, list4, 2)

	// Case 5: Pagination - Page 1, Size 1 for top-level
	list5Page1, total5, err5 := ListWorldviews(ctx, 0, "", 1, 1)
	assert.NoError(t, err5)
	assert.Equal(t, int64(3), total5)
	assert.Len(t, list5Page1, 1)

	// Case 6: Pagination - Page 2, Size 1 for top-level
	list5Page2, _, err5 := ListWorldviews(ctx, 0, "", 2, 1)
	assert.NoError(t, err5)
	assert.Len(t, list5Page2, 1)
	assert.NotEqual(t, list5Page1[0].ID, list5Page2[0].ID, "Page 1 and Page 2 items should be different")

	// Case 7: No results
	list7, total7, err7 := ListWorldviews(ctx, -1, "nonexistenttag", 1, 5)
	assert.NoError(t, err7)
	assert.Equal(t, int64(0), total7)
	assert.Len(t, list7, 0)
}
