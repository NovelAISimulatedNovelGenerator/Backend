package background


// Worldview 世界观实体，描述小说整体设定和宇宙观
// 包含名称、详细描述、软删除与时间戳
// 适用于大语言模型生成小说时提供大背景信息
// 表名: worldviews
//
// Worldview 世界观实体，支持主从分层结构
// 包含名称、描述、标签、父ID及子世界观
// 可用于树状世界观体系
//
type Worldview struct {
	ID          uint         // 主键ID
	Name        string      // 世界观名称
	Description string      // 世界观详细描述
	Tag         string      // 标签，多个标签用英文逗号分隔
	ParentID    uint        // 父世界观ID，0表示主世界观，否则为子世界观
	Children    []Worldview // 子世界观列表
}

// Rule 规则实体，描述世界观下的运行法则
// 包含名称、详细描述、所属世界观ID、软删除与时间戳
// 适用于大语言模型生成小说时提供规则设定
// 表名: rules
//
// Rule 规则实体，支持主从分层结构
// 包含名称、描述、标签、父ID及子规则
// 可用于树状规则体系
//
type Rule struct {
	ID           uint     // 主键ID
	WorldviewID  uint     // 所属世界观ID
	Name         string   // 规则名称
	Description  string   // 规则详细描述
	Tag          string   // 标签，多个标签用英文逗号分隔
	ParentID     uint     // 父规则ID，0表示主规则，否则为子规则
	Children     []Rule   // 子规则列表
}

// Background 背景实体，描述故事发生的具体设定
// 包含名称、详细描述、所属世界观ID、软删除与时间戳
// 适用于大语言模型生成小说时提供故事背景
// 表名: backgrounds
//
// Background 背景实体，支持主从分层结构
// 包含名称、描述、标签、父ID及子背景
// 可用于树状背景体系
//
type Background struct {
	ID           uint         // 主键ID
	WorldviewID  uint         // 所属世界观ID
	Name         string       // 背景名称
	Description  string       // 背景详细描述
	Tag          string       // 标签，多个标签用英文逗号分隔
	ParentID     uint         // 父背景ID，0表示主背景，否则为子背景
	Children     []Background // 子背景列表
}
