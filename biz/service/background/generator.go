package background

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"google.golang.org/protobuf/types/known/timestamppb"
	model "novelai/biz/model/background"
)

// NovelGeneratorOption 小说生成选项函数类型
// 允许使用函数选项模式配置小说生成过程
type NovelGeneratorOption func(*NovelGeneratorOptions) error

// NovelGeneratorOptions 小说生成配置
type NovelGeneratorOptions struct {
	// 世界观生成函数
	WorldviewGenerator func(context.Context) ([]*model.Worldview, error)
	// 规则生成函数，接收世界观作为参数
	RuleGenerator func(context.Context, []*model.Worldview) ([]*model.Rule, error)
	// 背景生成函数，接收世界观和规则作为参数
	BackgroundGenerator func(context.Context, []*model.Worldview, []*model.Rule) ([]*model.BackgroundInfo, error)
	// 小说后处理函数，可以在生成完成后修改小说
	PostProcessor func(context.Context, *NovelInfo) error
}

// NovelInfo 结构体，包含所有的世界观、规则和背景信息
type NovelInfo struct {
	ID          int64                      // 小说ID
	WorldViews  []*model.Worldview         // 世界观列表
	Rules       []*model.Rule              // 规则列表
	Backgrounds []*model.BackgroundInfo    // 背景列表
	CreatedAt   *timestamppb.Timestamp     // 创建时间
	UpdatedAt   *timestamppb.Timestamp     // 更新时间
}

// WithWorldviewGenerator 设置世界观生成函数
func WithWorldviewGenerator(gen func(context.Context) ([]*model.Worldview, error)) NovelGeneratorOption {
	return func(opts *NovelGeneratorOptions) error {
		if gen == nil {
			return errors.New("世界观生成函数不能为空")
		}
		opts.WorldviewGenerator = gen
		return nil
	}
}

// WithRuleGenerator 设置规则生成函数
func WithRuleGenerator(gen func(context.Context, []*model.Worldview) ([]*model.Rule, error)) NovelGeneratorOption {
	return func(opts *NovelGeneratorOptions) error {
		if gen == nil {
			return errors.New("规则生成函数不能为空")
		}
		opts.RuleGenerator = gen
		return nil
	}
}

// WithBackgroundGenerator 设置背景生成函数
func WithBackgroundGenerator(gen func(context.Context, []*model.Worldview, []*model.Rule) ([]*model.BackgroundInfo, error)) NovelGeneratorOption {
	return func(opts *NovelGeneratorOptions) error {
		if gen == nil {
			return errors.New("背景生成函数不能为空")
		}
		opts.BackgroundGenerator = gen
		return nil
	}
}

// WithPostProcessor 设置小说后处理函数
func WithPostProcessor(proc func(context.Context, *NovelInfo) error) NovelGeneratorOption {
	return func(opts *NovelGeneratorOptions) error {
		if proc == nil {
			return errors.New("后处理函数不能为空")
		}
		opts.PostProcessor = proc
		return nil
	}
}

// defaultWorldviewGenerator 默认世界观生成函数
func defaultWorldviewGenerator(ctx context.Context) ([]*model.Worldview, error) {
	// 使用 Hertz 的日志系统记录警告
	hlog.CtxWarnf(ctx, "正在使用默认世界观生成器，该生成器仅返回错误，请提供自定义WorldviewGenerator")
	// 返回错误，提示需要实现真实的生成逻辑
	return nil, errors.New("未实现世界观生成逻辑，请提供有效的WorldviewGenerator")
}

// defaultRuleGenerator 默认规则生成函数
func defaultRuleGenerator(ctx context.Context, worldviews []*model.Worldview) ([]*model.Rule, error) {
	// 使用 Hertz 的日志系统记录警告
	hlog.CtxWarnf(ctx, "正在使用默认规则生成器，该生成器仅返回错误，请提供自定义RuleGenerator")
	// 返回错误，提示需要实现真实的生成逻辑
	return nil, errors.New("未实现规则生成逻辑，请提供有效的RuleGenerator")
}

// defaultBackgroundGenerator 默认背景生成函数
func defaultBackgroundGenerator(ctx context.Context, worldviews []*model.Worldview, rules []*model.Rule) ([]*model.BackgroundInfo, error) {
	// 使用 Hertz 的日志系统记录警告
	hlog.CtxWarnf(ctx, "正在使用默认背景生成器，该生成器仅返回错误，请提供自定义BackgroundGenerator")
	// 返回错误，提示需要实现真实的生成逻辑
	return nil, errors.New("未实现背景生成逻辑，请提供有效的BackgroundGenerator")
}

// defaultPostProcessor 默认后处理函数
func defaultPostProcessor(ctx context.Context, novel *NovelInfo) error {
	// 使用 Hertz 的日志系统记录信息
	hlog.CtxInfof(ctx, "正在使用默认后处理器，该处理器不执行任何操作")
	// 后处理函数可以为空，因为它是可选的
	return nil
}

// defaultOptions 返回默认选项配置
func defaultOptions() *NovelGeneratorOptions {
	// 使用 Hertz 的日志系统记录警告
	hlog.Warnf("正在使用默认小说生成选项，默认选项使用的生成函数会抛出错误，请使用 WithWorldviewGenerator、WithRuleGenerator 等函数设置自定义生成器")

	return &NovelGeneratorOptions{
		WorldviewGenerator:  defaultWorldviewGenerator,
		RuleGenerator:       defaultRuleGenerator,
		BackgroundGenerator: defaultBackgroundGenerator,
		PostProcessor:       defaultPostProcessor,
	}
}

// generate 生成一个小说及其相关背景设定
// 参数:
// - ctx: 上下文，用于控制生成过程的取消和超时
// - options: 可变参数，用于自定义生成过程的各个方面
// 返回:
// - 生成的小说结构体
// - 如果生成过程出错，返回相应错误
func generate(ctx context.Context, options ...NovelGeneratorOption) (*NovelInfo, error) {
	// 检查上下文是否有效
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 初始化默认选项
	opts := defaultOptions()

	// 应用自定义选项
	for _, option := range options {
		if err := option(opts); err != nil {
			return nil, errors.New("应用选项失败: " + err.Error())
		}
	}

	// 创建小说结构
	novel := &NovelInfo{
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}

	// 生成世界观
	worldviews, err := opts.WorldviewGenerator(ctx)
	if err != nil {
		return nil, errors.New("生成世界观失败: " + err.Error())
	}
	novel.WorldViews = worldviews

	// 生成规则
	rules, err := opts.RuleGenerator(ctx, worldviews)
	if err != nil {
		return nil, errors.New("生成规则失败: " + err.Error())
	}
	novel.Rules = rules

	// 生成背景
	backgrounds, err := opts.BackgroundGenerator(ctx, worldviews, rules)
	if err != nil {
		return nil, errors.New("生成背景失败: " + err.Error())
	}
	novel.Backgrounds = backgrounds

	// 应用后处理
	if err := opts.PostProcessor(ctx, novel); err != nil {
		return nil, errors.New("后处理失败: " + err.Error())
	}

	return novel, nil
}
