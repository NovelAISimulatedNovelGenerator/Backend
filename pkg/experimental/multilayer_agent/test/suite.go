// package test 提供多层智能体系统的测试框架和用例
package test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"novelai/pkg/experimental/multilayer_agent/shared/model"
	agenttools "novelai/pkg/experimental/multilayer_agent/shared/tools"
)

// TestSuite 定义测试套件结构
type TestSuite struct {
	// 测试上下文
	Ctx context.Context

	// 测试配置
	Config TestConfig

	// 测试模型
	Model model.Model

	// 工具注册表
	ToolRegistry *agenttools.ToolRegistry

	// 测试用例集合
	TestCases []TestCase

	// 测试结果
	Results []TestResult

	// 互斥锁保护结果集
	mu sync.Mutex
}

// TestConfig 测试套件配置
type TestConfig struct {
	// 测试环境 (dev/prod)
	Environment string

	// 超时设置(秒)
	Timeout int

	// 是否输出详细日志
	VerboseLogging bool

	// 断言失败时是否继续测试
	ContinueOnFailure bool
}

// TestCase 定义测试用例接口
type TestCase interface {
	// 获取测试名称
	Name() string

	// 获取测试描述
	Description() string

	// 设置测试环境
	Setup(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) error

	// 执行测试
	Execute() error

	// 验证测试结果
	Verify() (bool, string)

	// 清理测试资源
	Cleanup()
}

// TestResult 测试结果结构
type TestResult struct {
	// 测试用例名称
	TestName string

	// 测试通过/失败
	Passed bool

	// 错误详情
	ErrorMessage string

	// 执行时间
	ExecutionTime time.Duration

	// 执行时间戳
	Timestamp time.Time
}

// DefaultTestConfig 返回默认测试配置
func DefaultTestConfig() TestConfig {
	return TestConfig{
		Environment:       "dev",
		Timeout:           30,
		VerboseLogging:    true,
		ContinueOnFailure: true,
	}
}

// NewTestSuite 创建新的测试套件
func NewTestSuite(ctx context.Context, model model.Model, registry *agenttools.ToolRegistry) *TestSuite {
	return &TestSuite{
		Ctx:          ctx,
		Config:       DefaultTestConfig(),
		Model:        model,
		ToolRegistry: registry,
		TestCases:    make([]TestCase, 0),
		Results:      make([]TestResult, 0),
	}
}

// AddTestCase 添加测试用例到套件
func (s *TestSuite) AddTestCase(tc TestCase) {
	s.TestCases = append(s.TestCases, tc)
}

// RunAll 运行所有测试用例
func (s *TestSuite) RunAll() []TestResult {
	hlog.Infof("开始执行测试套件，共 %d 个测试用例", len(s.TestCases))
	
	for _, tc := range s.TestCases {
		result := s.runTestCase(tc)
		s.mu.Lock()
		s.Results = append(s.Results, result)
		s.mu.Unlock()
		
		if !result.Passed && !s.Config.ContinueOnFailure {
			hlog.Warnf("测试用例 %s 失败，中止后续测试", tc.Name())
			break
		}
	}
	
	s.printSummary()
	return s.Results
}

// runTestCase 运行单个测试用例
func (s *TestSuite) runTestCase(tc TestCase) TestResult {
	result := TestResult{
		TestName:  tc.Name(),
		Timestamp: time.Now(),
	}
	
	hlog.Infof("\n===== 开始测试: %s =====", tc.Name())
	hlog.Infof("描述: %s", tc.Description())
	
	// 创建测试超时上下文
	ctx, cancel := context.WithTimeout(s.Ctx, time.Duration(s.Config.Timeout)*time.Second)
	defer cancel()
	
	startTime := time.Now()
	
	// 设置测试环境
	hlog.Infof("设置测试环境...")
	if err := tc.Setup(ctx, s.Model, s.ToolRegistry); err != nil {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("设置失败: %v", err)
		result.ExecutionTime = time.Since(startTime)
		hlog.Errorf("测试设置失败: %v", err)
		return result
	}
	
	// 执行测试
	hlog.Infof("执行测试...")
	err := tc.Execute()
	if err != nil {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("执行失败: %v", err)
		result.ExecutionTime = time.Since(startTime)
		hlog.Errorf("测试执行失败: %v", err)
		
		// 即使执行失败也尝试清理资源
		tc.Cleanup()
		return result
	}
	
	// 验证结果
	hlog.Infof("验证测试结果...")
	passed, message := tc.Verify()
	result.Passed = passed
	if !passed {
		result.ErrorMessage = fmt.Sprintf("验证失败: %s", message)
		hlog.Errorf("测试验证失败: %s", message)
	} else {
		hlog.Infof("测试验证通过")
	}
	
	// 清理资源
	hlog.Infof("清理测试资源...")
	tc.Cleanup()
	
	result.ExecutionTime = time.Since(startTime)
	statusText := "失败"
	if result.Passed {
		statusText = "通过"
	}
	hlog.Infof("测试用例 %s %s，耗时: %v", tc.Name(), statusText, result.ExecutionTime)
	
	return result
}

// printSummary 打印测试结果摘要
func (s *TestSuite) printSummary() {
	passed := 0
	failed := 0
	
	for _, result := range s.Results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}
	
	hlog.Infof("\n===== 测试套件执行完成 =====")
	hlog.Infof("总测试用例: %d", len(s.Results))
	hlog.Infof("通过: %d", passed)
	hlog.Infof("失败: %d", failed)
	
	if failed > 0 {
		hlog.Warnf("以下测试用例失败:")
		for _, result := range s.Results {
			if !result.Passed {
				hlog.Warnf("- %s: %s", result.TestName, result.ErrorMessage)
			}
		}
	}
}

// RunTests 运行集成测试的主入口函数
func RunTests(ctx context.Context, testModel model.Model, registry *agenttools.ToolRegistry) {
	hlog.Infof("开始多层智能体系统集成测试...")
	
	// 创建测试套件
	suite := NewTestSuite(ctx, testModel, registry)
	
	// 添加单元测试
	suite.AddTestCase(NewAgentInitializationTest())
	suite.AddTestCase(NewAgentMessageHandlingTest())
	suite.AddTestCase(NewAgentStateManagementTest())
	suite.AddTestCase(NewAgentToolUsageTest())
	
	// 添加组件测试
	suite.AddTestCase(NewOrchestratorLifecycleTest())
	suite.AddTestCase(NewOrchestratorAgentRegistryTest())
	suite.AddTestCase(NewOrchestratorMessageRoutingTest())
	
	// 添加集成测试
	suite.AddTestCase(NewMultiAgentCollaborationTest())
	suite.AddTestCase(NewToolChainIntegrationTest())
	
	// 添加系统测试
	suite.AddTestCase(NewNovelGenerationTest())
	
	// 运行所有测试用例
	suite.RunAll()
	
	hlog.Infof("多层智能体系统集成测试完成")
}
