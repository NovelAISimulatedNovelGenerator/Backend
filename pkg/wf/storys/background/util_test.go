package background

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"novelai/pkg/constants"
	"novelai/pkg/llm/deepseek"
)

func TestWorldviewStringAndParse(t *testing.T) {
	w := Worldview{
		ID:          1,
		Name:        "主世界观",
		Description: "这是主世界观描述",
		Tag:         "幻想,史诗",
		ParentID:    0,
	}
	str := w.String()
	w2, err := ParseWorldviewFromString(str)
	if err != nil {
		t.Fatalf("ParseWorldviewFromString failed: %v", err)
	}
	if !reflect.DeepEqual(w, w2) {
		t.Errorf("Worldview parse mismatch:\n原始: %#v\n还原: %#v", w, w2)
	}
}

func TestRuleStringAndParse(t *testing.T) {
	r := Rule{
		ID:          2,
		WorldviewID: 1,
		Name:        "魔法法则",
		Description: "魔法世界的核心规则",
		Tag:         "魔法,能量",
		ParentID:    0,
	}
	str := r.String()
	r2, err := ParseRuleFromString(str)
	if err != nil {
		t.Fatalf("ParseRuleFromString failed: %v", err)
	}
	r.WorldviewID = r2.WorldviewID // 保证一致性
	if !reflect.DeepEqual(r, r2) {
		t.Errorf("Rule parse mismatch:\n原始: %#v\n还原: %#v", r, r2)
	}
}

func TestBackgroundStringAndParse(t *testing.T) {
	b := Background{
		ID:          3,
		WorldviewID: 1,
		Name:        "远古大陆",
		Description: "故事发生的神秘大陆",
		Tag:         "大陆,远古",
		ParentID:    0,
	}
	str := b.String()
	b2, err := ParseBackgroundFromString(str)
	if err != nil {
		t.Fatalf("ParseBackgroundFromString failed: %v", err)
	}
	b.WorldviewID = b2.WorldviewID // 保证一致性
	if !reflect.DeepEqual(b, b2) {
		t.Errorf("Background parse mismatch:\n原始: %#v\n还原: %#v", b, b2)
	}
}

func TestActualCreation(t *testing.T) {
	prompt := "你是一个小说背景生成助手，请生成一个主世界观，包括名称、描述、标签。请严格按照如下 JSON 格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}不要输出除 JSON 以外的内容。"
	apiKey := "sk-35defd1d4a64457f88e849454d21f17f"
	messages := []deepseek.Message{
		{
			Role:    "system",
			Content: prompt,
		},
	}
	if apiKey == "" {
		t.Skip("请先设置环境变量 DEEPSEEK_API_KEY")
		return
	}

	// 创建 DeepSeek 客户端，baseurl 只提供基础域名
	config := deepseek.DefaultConfig(apiKey).WithBaseURL("https://api.deepseek.com")
	client, err := deepseek.NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("创建客户端错误: %v", err)
	}
	chatcompReq := &deepseek.ChatRequest{
		Model:       constants.DeepSeekChat,
		Messages:    messages,
		MaxTokens:   100,
		Temperature: 0.7,
		ResponseFormat: deepseek.ResponseFormat{
			Type: "json_object",
		},
	}

	ctx := context.Background()

	chatcompResp, err := client.ChatCompletion(ctx, chatcompReq)
	if err != nil {
		t.Fatalf("ChatComp 错误: %v", err)
	}

	// 检测 chatcomp 回答能否被json序列化
	// 逐步安全提取 chatcompResp["choices"][0]["message"]["content"]
	choicesRaw, ok := chatcompResp["choices"]
	if !ok {
		t.Fatalf("chatcompResp 缺少 choices 字段")
	}
	choicesArr, ok := choicesRaw.([]interface{})
	if !ok || len(choicesArr) == 0 {
		t.Fatalf("choices 字段类型错误或为空")
	}
	choiceMap, ok := choicesArr[0].(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0] 类型错误")
	}
	messageRaw, ok := choiceMap["message"]
	if !ok {
		t.Fatalf("choices[0] 缺少 message 字段")
	}
	messageMap, ok := messageRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("choices[0].message 类型错误")
	}
	contentRaw, ok := messageMap["content"]
	if !ok {
		t.Fatalf("choices[0].message 缺少 content 字段")
	}
	contentStr, ok := contentRaw.(string)
	if !ok {
		t.Fatalf("choices[0].message.content 类型不是 string")
	}

	var jsonResp map[string]string
	if err := json.Unmarshal([]byte(contentStr), &jsonResp); err != nil {
		t.Errorf("无法解析JSON响应: %v, 原始响应: %s", err, contentStr)
	}

	// 验证必要字段
	requiredFields := []string{"name", "description", "tag"}
	for _, field := range requiredFields {
		if _, exists := jsonResp[field]; !exists {
			t.Errorf("JSON响应缺少必要字段: %s", field)
		}
	}
}
