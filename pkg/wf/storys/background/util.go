package background

import (
	"fmt"
	"regexp"
	"strconv"
	"errors"
)

// Worldview 的字符串表示
func (w Worldview) String() string {
	return fmt.Sprintf("Worldview[ID=%d, Name=%s, Description=%s, Tag=%s, ParentID=%d]", w.ID, w.Name, w.Description, w.Tag, w.ParentID)
}

// ParseWorldviewFromString 将 String() 输出还原为 Worldview 结构体
func ParseWorldviewFromString(s string) (Worldview, error) {
	pattern := `Worldview\[ID=(\d+), Name=(.*?), Description=(.*?), Tag=(.*?), ParentID=(\d+)\]`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 6 {
		return Worldview{}, errors.New("string format not match Worldview pattern")
	}
	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Worldview{}, err
	}
	parentID, err := strconv.ParseUint(matches[5], 10, 64)
	if err != nil {
		return Worldview{}, err
	}
	return Worldview{
		ID:          uint(id),
		Name:        matches[2],
		Description: matches[3],
		Tag:         matches[4],
		ParentID:    uint(parentID),
	}, nil
}

// Rule 的字符串表示
func (r Rule) String() string {
	return fmt.Sprintf("Rule[ID=%d, Name=%s, Description=%s, Tag=%s, ParentID=%d, WorldviewID=%d]", r.ID, r.Name, r.Description, r.Tag, r.ParentID, r.WorldviewID)
}

// ParseRuleFromString 将 String() 输出还原为 Rule 结构体
func ParseRuleFromString(s string) (Rule, error) {
	pattern := `Rule\[ID=(\d+), Name=(.*?), Description=(.*?), Tag=(.*?), ParentID=(\d+), WorldviewID=(\d+)\]`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 7 {
		return Rule{}, errors.New("string format not match Rule pattern")
	}
	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Rule{}, err
	}
	parentID, err := strconv.ParseUint(matches[5], 10, 64)
	if err != nil {
		return Rule{}, err
	}
	worldviewID, err := strconv.ParseUint(matches[6], 10, 64)
	if err != nil {
		return Rule{}, err
	}
	return Rule{
		ID:           uint(id),
		Name:         matches[2],
		Description:  matches[3],
		Tag:          matches[4],
		ParentID:     uint(parentID),
		WorldviewID:  uint(worldviewID),
	}, nil
}

// Background 的字符串表示
func (b Background) String() string {
	return fmt.Sprintf("Background[ID=%d, Name=%s, Description=%s, Tag=%s, ParentID=%d, WorldviewID=%d]", b.ID, b.Name, b.Description, b.Tag, b.ParentID, b.WorldviewID)
}

// ParseBackgroundFromString 将 String() 输出还原为 Background 结构体
func ParseBackgroundFromString(s string) (Background, error) {
	pattern := `Background\[ID=(\d+), Name=(.*?), Description=(.*?), Tag=(.*?), ParentID=(\d+), WorldviewID=(\d+)\]`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 7 {
		return Background{}, errors.New("string format not match Background pattern")
	}
	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Background{}, err
	}
	parentID, err := strconv.ParseUint(matches[5], 10, 64)
	if err != nil {
		return Background{}, err
	}
	worldviewID, err := strconv.ParseUint(matches[6], 10, 64)
	if err != nil {
		return Background{}, err
	}
	return Background{
		ID:           uint(id),
		Name:         matches[2],
		Description:  matches[3],
		Tag:          matches[4],
		ParentID:     uint(parentID),
		WorldviewID:  uint(worldviewID),
	}, nil
}
