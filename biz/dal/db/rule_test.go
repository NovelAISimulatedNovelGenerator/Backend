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

// setupTestDBForRule initializes the database for Rule tests.
// It also migrates Worldview as Rule depends on it.
func setupTestDBForRule(t *testing.T) {
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database for Rule")

	// Rule depends on Worldview, so migrate both
	err = DB.AutoMigrate(&Worldview{}, &Rule{})
	assert.NoError(t, err, "Failed to migrate Rule and Worldview tables")

	// Clean tables before each specific test suite
	DB.Exec("DELETE FROM " + constants.TableNameRule)
	DB.Exec("DELETE FROM " + constants.TableNameWorldview) // Clean dependency table too
}

// createTestRule creates a test rule instance and persists it.
// Requires a worldviewID for the rule.
func createTestRule(t *testing.T, nameSuffix string, worldviewID int64, parentID int64, tag string) *Rule {
	ctx := context.Background()
	r := &Rule{
		Name:        "Test Rule " + nameSuffix,
		Description: "Desc for rule " + nameSuffix,
		WorldviewID: worldviewID,
		Tag:         tag,
		ParentID:    parentID,
	}
	id, err := CreateRule(ctx, r)
	assert.NoError(t, err, "CreateRule failed for "+nameSuffix)
	assert.Greater(t, id, int64(0))

	createdRule, err := GetRuleByID(ctx, id)
	assert.NoError(t, err, "GetRuleByID failed after create for "+nameSuffix)
	assert.NotNil(t, createdRule)
	return createdRule
}

func TestCRUDRule(t *testing.T) {
	setupTestDBForRule(t)
	ctx := context.Background()

	// Create a prerequisite Worldview
	testWv := createTestWorldview(t, "ForRule", 0, "wv_for_rule")
	assert.NotNil(t, testWv, "Test Worldview for Rule creation cannot be nil")

	// Test Create
	r1 := &Rule{
		Name:        "Primary Rule",
		Description: "This is the main rule.",
		WorldviewID: testWv.ID,
		Tag:         "primary,core",
		ParentID:    0,
	}
	id1, err := CreateRule(ctx, r1)
	assert.NoError(t, err)
	assert.Greater(t, id1, int64(0))

	fetchedR1, err := GetRuleByID(ctx, id1)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedR1)
	assert.Equal(t, r1.Name, fetchedR1.Name)
	assert.Equal(t, testWv.ID, fetchedR1.WorldviewID)

	// Test Get Non-Existing
	_, err = GetRuleByID(ctx, 99999)
	assert.ErrorIs(t, err, ErrRuleNotFound)

	// Test Update
	updates := map[string]interface{}{
		"name":        "Updated Rule Name",
		"description": "Updated rule desc.",
		"tag":         "updated,new_rule_tag",
	}
	oldUpdatedAt := fetchedR1.UpdatedAt
	time.Sleep(1100 * time.Millisecond) // Ensure UpdatedAt changes (second precision for int64 timestamp)
	err = UpdateRule(ctx, id1, updates)
	assert.NoError(t, err)

	updatedRule, err := GetRuleByID(ctx, id1)
	assert.NoError(t, err)
	assert.Equal(t, updates["name"], updatedRule.Name)
	assert.Equal(t, updates["tag"], updatedRule.Tag)
	assert.Greater(t, updatedRule.UpdatedAt, oldUpdatedAt, "UpdatedAt timestamp (%v) should be greater than oldUpdatedAt (%v) after update", updatedRule.UpdatedAt, oldUpdatedAt)

	// Test Update Non-Existing
	err = UpdateRule(ctx, 99999, updates)
	assert.Error(t, err) // Exact error can be ErrRuleNotFound or ErrUpdateRuleFailed based on GORM behavior

	// Test Delete
	err = DeleteRule(ctx, id1)
	assert.NoError(t, err)
	_, err = GetRuleByID(ctx, id1)
	assert.ErrorIs(t, err, ErrRuleNotFound)

	// Test Delete Non-Existing
	err = DeleteRule(ctx, 99999)
	assert.ErrorIs(t, err, ErrRuleNotFound)
}

func TestListRules(t *testing.T) {
	setupTestDBForRule(t)
	ctx := context.Background()

	wv1 := createTestWorldview(t, "Wv1ForRuleList", 0, "wv1")
	wv2 := createTestWorldview(t, "Wv2ForRuleList", 0, "wv2")

	ruleParentWv1 := createTestRule(t, "ParentRuleWv1", wv1.ID, 0, "parent,wv1_rule")
	_ = createTestRule(t, "Child1RuleWv1", wv1.ID, ruleParentWv1.ID, "child,tagX")
	_ = createTestRule(t, "Child2RuleWv1", wv1.ID, ruleParentWv1.ID, "child,tagY")

	_ = createTestRule(t, "TopRuleWv2", wv2.ID, 0, "top,tagX")
	_ = createTestRule(t, "AnotherRuleWv1", wv1.ID, 0, "general,tagZ")

	// Case 1: List rules for wv1
	list1, total1, err1 := ListRules(ctx, wv1.ID, -1, "", 1, 10)
	assert.NoError(t, err1)
	assert.Equal(t, int64(4), total1) // ParentRuleWv1, Child1, Child2, AnotherRuleWv1
	assert.Len(t, list1, 4)

	// Case 2: List children of ruleParentWv1 (parentIDFilter = ruleParentWv1.ID)
	list2, total2, err2 := ListRules(ctx, wv1.ID, ruleParentWv1.ID, "", 1, 10)
	assert.NoError(t, err2)
	assert.Equal(t, int64(2), total2)
	assert.Len(t, list2, 2)

	// Case 3: Filter by tag "tagX" (any worldview, any parent)
	list3, total3, err3 := ListRules(ctx, 0, -1, "tagX", 1, 10)
	assert.NoError(t, err3)
	assert.Equal(t, int64(2), total3) // Child1RuleWv1 (wv1), TopRuleWv2 (wv2)
	assert.Len(t, list3, 2)

	// Case 4: Filter by tag "child" for wv1
	list4, total4, err4 := ListRules(ctx, wv1.ID, -1, "child", 1, 10)
	assert.NoError(t, err4)
	assert.Equal(t, int64(2), total4)
	assert.Len(t, list4, 2)

	// Case 5: Pagination for wv1 rules - Page 1, Size 1
	list5Page1, total5, err5 := ListRules(ctx, wv1.ID, -1, "", 1, 2) // Expecting 4 rules for wv1
	assert.NoError(t, err5)
	assert.Equal(t, int64(4), total5)
	assert.Len(t, list5Page1, 2)

	// Case 6: Pagination for wv1 rules - Page 2, Size 2
	list5Page2, _, err5 := ListRules(ctx, wv1.ID, -1, "", 2, 2)
	assert.NoError(t, err5)
	assert.Len(t, list5Page2, 2)
	assert.NotEqual(t, list5Page1[0].ID, list5Page2[0].ID, "Page 1 and Page 2 items should be different")

	// Case 7: No results for a specific worldview with non-matching tag
	list7, total7, err7 := ListRules(ctx, wv1.ID, -1, "nonexistenttag", 1, 10)
	assert.NoError(t, err7)
	assert.Equal(t, int64(0), total7)
	assert.Len(t, list7, 0)
}
