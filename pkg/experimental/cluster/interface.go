package cluster

import (
	"context"
)

// DistanceFunc 定义了两个向量之间距离计算的函数类型
type DistanceFunc func(a, b []float64) float64

// Clusterer 聚类算法的抽象接口
type Clusterer interface {
	// Fit 执行聚类算法
	Fit(ctx context.Context, data [][]float64) error
	
	// Predict 根据训练好的模型预测新数据点所属的聚类
	Predict(ctx context.Context, point []float64) (int, error)
	
	// GetClusters 返回所有聚类及其包含的数据点
	GetClusters(ctx context.Context) ([][]int, error)
	
	// GetCentroids 返回所有聚类的中心点
	GetCentroids(ctx context.Context) ([][]float64, error)
}

// ClusterParams 聚类算法的通用参数
type ClusterParams struct {
	// K 指定聚类的数量
	K int
	
	// MaxIterations 指定最大迭代次数
	MaxIterations int
	
	// DistanceFunc 指定距离计算函数
	DistanceFunc DistanceFunc
	
	// Tolerance 指定收敛阈值
	Tolerance float64
}
