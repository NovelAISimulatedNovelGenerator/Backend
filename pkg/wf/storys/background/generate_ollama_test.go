package background

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ollama/ollama/api"
)

// TestGenerateWithOllama 测试使用Ollama生成故事背景
func TestGenerateWithOllama(t *testing.T) {
	// 跳过实际联网测试，除非显式启用
	if testing.Short() {
		t.Skip("跳过需要Ollama API的测试")
	}

	// 创建客户端
	client, err := api.ClientFromEnvironment()
	if err != nil {
		t.Fatalf("创建Ollama客户端失败: %v", err)
	}

	// 创建上下文
	ctx := context.Background()

	// 测试使用Ollama生成规则数据
	t.Run("测试生成规则", func(t *testing.T) {
		// 创建生成规则的提示词
		prompt := "你是一个小说规则生成助手，请生成一个世界规则，包括名称、描述、标签。请严格按照如下JSON格式输出：{\"name\": \"\", \"description\": \"\", \"tag\": \"\"}不要输出除JSON以外的内容。"
		
		// 创建请求
		req := &api.GenerateRequest{
			Model:  "deepseek-r1:14b",
			Prompt: prompt,
			Stream: new(bool), // 设置非流式输出
			Format: json.RawMessage(`"json"`),
		}

		var jsonResponse string
		// 定义响应处理函数
		respFunc := func(resp api.GenerateResponse) error {
			jsonResponse = resp.Response
			t.Logf("原始Ollama响应: %s", jsonResponse)
			return nil
		}

		// 调用API
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			t.Fatal(err)
		}

		// 解析JSON响应
		var result struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Tag         string `json:"tag"`
		}

		err = json.Unmarshal([]byte(jsonResponse), &result)
		if err != nil {
			t.Fatalf("JSON解析失败: %v", err)
		}
		t.Logf("解析后的数据: name=%s, description=%s, tag=%s", result.Name, result.Description, result.Tag)

		// 创建Rule对象
		rule := Rule{
			ID:          1,
			WorldviewID: 1,
			Name:        result.Name,
			Description: result.Description,
			Tag:         result.Tag,
			ParentID:    0,
		}
		t.Logf("创建的Rule对象: %+v", rule)

		// 调用String()方法
		ruleStr := rule.String()
		t.Logf("通过String()方法输出: %s", ruleStr)

		// 解析回结构体
		parsedRule, err := ParseRuleFromString(ruleStr)
		if err != nil {
			t.Fatalf("解析Rule字符串失败: %v", err)
		}
		t.Logf("解析后的Rule对象: %+v", parsedRule)

		// 验证一致性
		if parsedRule.ID != rule.ID ||
			parsedRule.Name != rule.Name ||
			parsedRule.Description != rule.Description ||
			parsedRule.Tag != rule.Tag ||
			parsedRule.ParentID != rule.ParentID ||
			parsedRule.WorldviewID != rule.WorldviewID {
			t.Errorf("转换前后数据不一致!\n原始: %+v\n解析后: %+v", rule, parsedRule)
		} else {
			t.Log("✅ 转换流程测试通过，数据一致性验证成功!")
		}
	})

	// 测试生成完整故事
	t.Run("测试生成故事", func(t *testing.T) {
		// 定义世界观生成器
		worldviewGenerator := func(ctx context.Context) ([]Worldview, error) {
			// 创建世界观对象(使用固定值而非API调用，减少不必要的网络请求)
			worldview := Worldview{
				ID:          1,
				Name:        "科技纪元",
				Description: "一个高度发达的科技世界",
				Tag:         "科幻,未来",
				ParentID:    0,
			}
			return []Worldview{worldview}, nil
		}

		// 定义规则生成器
		ruleGenerator := func(ctx context.Context, worldviews []Worldview) ([]Rule, error) {
			// 创建规则对象(使用固定值)
			rule := Rule{
				ID:          1,
				WorldviewID: 1,
				Name:        "能源法则",
				Description: "所有科技产品必须使用可再生能源",
				Tag:         "能源,科技",
				ParentID:    0,
			}
			return []Rule{rule}, nil
		}

		// 使用Ollama生成背景
		backgroundGenerator := func(ctx context.Context, worldviews []Worldview, rules []Rule) ([]Background, error) {
			// 创建生成背景的提示词
			prompt := "请根据以下信息生成一个故事背景，包括名称、描述、标签。使用JSON格式输出。\n"
			prompt += "世界观: 科技纪元\n"
			prompt += "世界观描述: 一个高度发达的科技世界\n"
			prompt += "规则: 能源法则\n"
			prompt += "规则描述: 所有科技产品必须使用可再生能源\n"
			
			// 创建请求
			req := &api.GenerateRequest{
				Model:  "deepseek-r1:14b",
				Prompt: prompt,
				Stream: new(bool),
				Format: json.RawMessage(`"json"`),
			}

			var jsonResponse string
			// 响应处理函数
			respFunc := func(resp api.GenerateResponse) error {
				jsonResponse = resp.Response
				t.Logf("背景生成响应: %s", jsonResponse)
				return nil
			}

			// 调用API
			err := client.Generate(ctx, req, respFunc)
			if err != nil {
				return nil, err
			}

			// 解析响应创建背景对象(简化处理，直接使用固定值)
			background := Background{
				ID:          1,
				WorldviewID: 1,
				Name:        "绿色城市",
				Description: "一座完全使用可再生能源的未来城市",
				Tag:         "城市,绿色",
				ParentID:    0,
			}

			return []Background{background}, nil
		}

		// 使用Generate函数生成故事
		story, err := Generate(ctx,
			WithWorldviewGenerator(worldviewGenerator),
			WithRuleGenerator(ruleGenerator),
			WithBackgroundGenerator(backgroundGenerator),
		)
		
		if err != nil {
			t.Fatalf("生成故事失败: %v", err)
		}

		// 验证生成的故事
		t.Logf("生成的故事: %d个世界观, %d个规则, %d个背景", 
			len(story.WorldViews), len(story.Rules), len(story.Backgrounds))
		
		if len(story.WorldViews) == 0 {
			t.Error("故事没有世界观")
		}
		if len(story.Rules) == 0 {
			t.Error("故事没有规则")
		}
		if len(story.Backgrounds) == 0 {
			t.Error("故事没有背景")
		}
		
		t.Log("✅ Generate函数测试成功")
	})
}
