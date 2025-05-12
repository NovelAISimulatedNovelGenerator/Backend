package db

import (
	"context"
	"testing"
	"time"

	"novelai/pkg/constants"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDBForBackgroundInfo initializes the database for BackgroundInfo tests.
// It also migrates Worldview as BackgroundInfo depends on it.
func setupTestDBForBackgroundInfo(t *testing.T) {
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database for BackgroundInfo")

	// BackgroundInfo depends on Worldview, so migrate both
	err = DB.AutoMigrate(&Worldview{}, &BackgroundInfo{})
	assert.NoError(t, err, "Failed to migrate BackgroundInfo and Worldview tables")

	// Clean tables before each specific test suite
	DB.Exec("DELETE FROM " + constants.TableNameBackgroundInfo)
	DB.Exec("DELETE FROM " + constants.TableNameWorldview) // Clean dependency table too
}

// createTestBackgroundInfo creates a test background_info instance and persists it.
// Requires a worldviewID.
func createTestBackgroundInfo(t *testing.T, nameSuffix string, worldviewID int64, parentID int64, tag string) *BackgroundInfo {
	ctx := context.Background()
	bi := &BackgroundInfo{
		Name:        "Test Background " + nameSuffix,
		Description: "Desc for background " + nameSuffix,
		WorldviewID: worldviewID,
		Tag:         tag,
		ParentID:    parentID,
	}
	id, err := CreateBackgroundInfo(ctx, bi)
	assert.NoError(t, err, "CreateBackgroundInfo failed for "+nameSuffix)
	assert.Greater(t, id, int64(0))

	createdBi, err := GetBackgroundInfoByID(ctx, id)
	assert.NoError(t, err, "GetBackgroundInfoByID failed after create for "+nameSuffix)
	assert.NotNil(t, createdBi)
	return createdBi
}

func TestCRUDBackgroundInfo(t *testing.T) {
	setupTestDBForBackgroundInfo(t)
	ctx := context.Background()

	// Create a prerequisite Worldview
	testWv := createTestWorldview(t, "ForBgInfo", 0, "wv_for_bg") // Using helper from worldview_test.go (assuming same package)
	assert.NotNil(t, testWv, "Test Worldview for BackgroundInfo creation cannot be nil")

	// Test Create
	bi1 := &BackgroundInfo{
		Name:        "Main Character Backstory",
		Description: "Details about the protagonist.",
		WorldviewID: testWv.ID,
		Tag:         "character,protagonist",
		ParentID:    0,
	}
	id1, err := CreateBackgroundInfo(ctx, bi1)
	assert.NoError(t, err)
	assert.Greater(t, id1, int64(0))

	fetchedBi1, err := GetBackgroundInfoByID(ctx, id1)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedBi1)
	assert.Equal(t, bi1.Name, fetchedBi1.Name)
	assert.Equal(t, testWv.ID, fetchedBi1.WorldviewID)

	// Test Get Non-Existing
	_, err = GetBackgroundInfoByID(ctx, 99999)
	assert.ErrorIs(t, err, ErrBackgroundInfoNotFound)

	// Test Update
	updates := map[string]interface{}{
		"name":        "Updated Backstory Name",
		"description": "More updated details.",
		"tag":         "character,updated_tag",
	}
	oldUpdatedAt := fetchedBi1.UpdatedAt
	time.Sleep(1100 * time.Millisecond) // Ensure UpdatedAt changes (second precision for int64 timestamp)
	err = UpdateBackgroundInfo(ctx, id1, updates)
	assert.NoError(t, err)

	updatedBi, err := GetBackgroundInfoByID(ctx, id1)
	assert.NoError(t, err)
	assert.Equal(t, updates["name"], updatedBi.Name)
	assert.Equal(t, updates["tag"], updatedBi.Tag)
	assert.Greater(t, updatedBi.UpdatedAt, oldUpdatedAt, "UpdatedAt timestamp (%v) should be greater than oldUpdatedAt (%v) after update", updatedBi.UpdatedAt, oldUpdatedAt)

	// Test Update Non-Existing
	err = UpdateBackgroundInfo(ctx, 99999, updates)
	assert.Error(t, err) 

	// Test Delete
	err = DeleteBackgroundInfo(ctx, id1)
	assert.NoError(t, err)
	_, err = GetBackgroundInfoByID(ctx, id1)
	assert.ErrorIs(t, err, ErrBackgroundInfoNotFound)

	// Test Delete Non-Existing
	err = DeleteBackgroundInfo(ctx, 99999)
	assert.ErrorIs(t, err, ErrBackgroundInfoNotFound)
}

func TestListBackgroundInfos(t *testing.T) {
	setupTestDBForBackgroundInfo(t)
	ctx := context.Background()

	wv1 := createTestWorldview(t, "Wv1ForBgList", 0, "wv1_bg_list")
	wv2 := createTestWorldview(t, "Wv2ForBgList", 0, "wv2_bg_list")

	bgParentWv1 := createTestBackgroundInfo(t, "ParentBgWv1", wv1.ID, 0, "parent_bg,wv1_info")
	_ = createTestBackgroundInfo(t, "Child1BgWv1", wv1.ID, bgParentWv1.ID, "child_bg,tagP")
	_ = createTestBackgroundInfo(t, "Child2BgWv1", wv1.ID, bgParentWv1.ID, "child_bg,tagQ")

	_ = createTestBackgroundInfo(t, "TopBgWv2", wv2.ID, 0, "top_bg,tagP")
	_ = createTestBackgroundInfo(t, "AnotherBgWv1", wv1.ID, 0, "general_bg,tagR")

	// Case 1: List background infos for wv1
	list1, total1, err1 := ListBackgroundInfos(ctx, wv1.ID, -1, "", 1, 10)
	assert.NoError(t, err1)
	assert.Equal(t, int64(4), total1) // ParentBgWv1, Child1, Child2, AnotherBgWv1
	assert.Len(t, list1, 4)

	// Case 2: List children of bgParentWv1 (parentIDFilter = bgParentWv1.ID)
	list2, total2, err2 := ListBackgroundInfos(ctx, wv1.ID, bgParentWv1.ID, "", 1, 10)
	assert.NoError(t, err2)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, list2, 2)

	// Case 3: Filter by tag "tagP" (any worldview, any parent)
	list3, total3, err3 := ListBackgroundInfos(ctx, 0, -1, "tagP", 1, 10)
	assert.NoError(t, err3)
	assert.Equal(t, int64(2), total3) // Child1BgWv1 (wv1), TopBgWv2 (wv2)
	assert.Len(t, list3, 2)

	// Case 4: Filter by tag "child_bg" for wv1
	list4, total4, err4 := ListBackgroundInfos(ctx, wv1.ID, -1, "child_bg", 1, 10)
	assert.NoError(t, err4)
	assert.Equal(t, int64(2), total4)
	assert.Len(t, list4, 2)

	// Case 5: Pagination for wv1 background infos - Page 1, Size 2
	list5Page1, total5, err5 := ListBackgroundInfos(ctx, wv1.ID, -1, "", 1, 2) // Expecting 4 for wv1
	assert.NoError(t, err5)
	assert.Equal(t, int64(4), total5)
	assert.Len(t, list5Page1, 2)

	// Case 6: Pagination for wv1 background infos - Page 2, Size 2
	list5Page2, _, err5 := ListBackgroundInfos(ctx, wv1.ID, -1, "", 2, 2)
	assert.NoError(t, err5)
	assert.Len(t, list5Page2, 2)
	assert.NotEqual(t, list5Page1[0].ID, list5Page2[0].ID)

	// Case 7: No results for a specific worldview with non-matching tag
	list7, total7, err7 := ListBackgroundInfos(ctx, wv1.ID, -1, "nonexistenttag_bg", 1, 10)
	assert.NoError(t, err7)
	assert.Equal(t, int64(0), total7)
	assert.Len(t, list7, 0)
}
