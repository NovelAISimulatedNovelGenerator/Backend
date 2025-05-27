// package test 提供多层智能体系统测试框架的主入口
package test

import (
	"context"

	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// RunMultiAgentSystemTests 执行多层智能体系统集成测试
// 参数:
//   - ctx: 上下文
//   - testModel: 测试使用的模型实例
//   - registry: 工具注册表
//
// 该函数是多层智能体系统测试的主入口点，会执行完整的测试套件
func RunMultiAgentSystemTests(ctx context.Context, testModel model.Model, registry *agenttools.ToolRegistry) {
	hlog.Infof("开始执行多层智能体系统集成测试...")

	// 运行测试套件
	RunTests(ctx, testModel, registry)

	hlog.Infof("多层智能体系统集成测试完成")
}
