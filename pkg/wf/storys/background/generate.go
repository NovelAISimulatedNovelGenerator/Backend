package background

import (
	"context"
	"errors"
)

// StoryOption 故事生成选项函数类型
// 允许使用函数选项模式配置故事生成过程
type StoryOption func(*StoryOptions) error

// StoryOptions 故事生成配置
type StoryOptions struct {
	// 世界观生成函数
	WorldviewGenerator func(context.Context) ([]Worldview, error)
	// 规则生成函数，接收世界观作为参数
	RuleGenerator func(context.Context, []Worldview) ([]Rule, error)
	// 背景生成函数，接收世界观和规则作为参数
	BackgroundGenerator func(context.Context, []Worldview, []Rule) ([]Background, error)
	// 故事后处理函数，可以在生成完成后修改故事
	PostProcessor func(context.Context, *Story) error
}

// WithWorldviewGenerator 设置世界观生成函数
func WithWorldviewGenerator(gen func(context.Context) ([]Worldview, error)) StoryOption {
	return func(opts *StoryOptions) error {
		if gen == nil {
			return errors.New("世界观生成函数不能为空")
		}
		opts.WorldviewGenerator = gen
		return nil
	}
}

// WithRuleGenerator 设置规则生成函数
func WithRuleGenerator(gen func(context.Context, []Worldview) ([]Rule, error)) StoryOption {
	return func(opts *StoryOptions) error {
		if gen == nil {
			return errors.New("规则生成函数不能为空")
		}
		opts.RuleGenerator = gen
		return nil
	}
}

// WithBackgroundGenerator 设置背景生成函数
func WithBackgroundGenerator(gen func(context.Context, []Worldview, []Rule) ([]Background, error)) StoryOption {
	return func(opts *StoryOptions) error {
		if gen == nil {
			return errors.New("背景生成函数不能为空")
		}
		opts.BackgroundGenerator = gen
		return nil
	}
}

// WithPostProcessor 设置故事后处理函数
func WithPostProcessor(proc func(context.Context, *Story) error) StoryOption {
	return func(opts *StoryOptions) error {
		if proc == nil {
			return errors.New("后处理函数不能为空")
		}
		opts.PostProcessor = proc
		return nil
	}
}

// defaultWorldviewGenerator 默认世界观生成函数
func defaultWorldviewGenerator(ctx context.Context) ([]Worldview, error) {
	// 返回空世界观列表，实际应用中可替换为真实生成逻辑
	return []Worldview{}, nil
}

// defaultRuleGenerator 默认规则生成函数
func defaultRuleGenerator(ctx context.Context, worldviews []Worldview) ([]Rule, error) {
	// 返回空规则列表，实际应用中可替换为真实生成逻辑
	return []Rule{}, nil
}

// defaultBackgroundGenerator 默认背景生成函数
func defaultBackgroundGenerator(ctx context.Context, worldviews []Worldview, rules []Rule) ([]Background, error) {
	// 返回空背景列表，实际应用中可替换为真实生成逻辑
	return []Background{}, nil
}

// defaultPostProcessor 默认后处理函数
func defaultPostProcessor(ctx context.Context, story *Story) error {
	// 默认不做任何处理，实际应用中可替换为真实处理逻辑
	return nil
}

// defaultOptions 返回默认选项配置
func defaultOptions() *StoryOptions {
	return &StoryOptions{
		WorldviewGenerator:  defaultWorldviewGenerator,
		RuleGenerator:       defaultRuleGenerator,
		BackgroundGenerator: defaultBackgroundGenerator,
		PostProcessor:       defaultPostProcessor,
	}
}

// Generate 生成一个故事及其相关背景设定
// 参数:
// - ctx: 上下文，用于控制生成过程的取消和超时
// - options: 可变参数，用于自定义生成过程的各个方面
// 返回:
// - 生成的故事结构体
// - 如果生成过程出错，返回相应错误
func Generate(ctx context.Context, options ...StoryOption) (Story, error) {
	// 检查上下文是否有效
	if ctx.Err() != nil {
		return Story{}, ctx.Err()
	}

	// 初始化默认选项
	opts := defaultOptions()

	// 应用自定义选项
	for _, option := range options {
		if err := option(opts); err != nil {
			return Story{}, errors.New("应用选项失败: " + err.Error())
		}
	}

	// 创建故事结构
	story := Story{}

	// 生成世界观
	worldviews, err := opts.WorldviewGenerator(ctx)
	if err != nil {
		return Story{}, errors.New("生成世界观失败: " + err.Error())
	}
	story.WorldViews = worldviews

	// 生成规则
	rules, err := opts.RuleGenerator(ctx, worldviews)
	if err != nil {
		return Story{}, errors.New("生成规则失败: " + err.Error())
	}
	story.Rules = rules

	// 生成背景
	backgrounds, err := opts.BackgroundGenerator(ctx, worldviews, rules)
	if err != nil {
		return Story{}, errors.New("生成背景失败: " + err.Error())
	}
	story.Backgrounds = backgrounds

	// 应用后处理
	if err := opts.PostProcessor(ctx, &story); err != nil {
		return Story{}, errors.New("后处理失败: " + err.Error())
	}

	return story, nil
}
