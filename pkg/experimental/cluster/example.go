package cluster

import (
	"context"
	"math/rand"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// ExampleData 生成示例数据用于聚类演示
func ExampleData(numClusters, pointsPerCluster int, noise float64) [][]float64 {
	rand.Seed(time.Now().UnixNano())
	
	// 总数据点数量
	totalPoints := numClusters * pointsPerCluster
	
	// 初始化数据集
	data := make([][]float64, totalPoints)
	
	// 生成每个聚类的中心点位置
	centers := make([][]float64, numClusters)
	for i := range centers {
		// 生成2D坐标，中心点相距较远
		centers[i] = []float64{
			rand.Float64()*10 + float64(i*10),
			rand.Float64()*10 + float64(i*10),
		}
	}
	
	// 生成每个聚类的数据点
	for i := 0; i < numClusters; i++ {
		for j := 0; j < pointsPerCluster; j++ {
			idx := i*pointsPerCluster + j
			
			// 初始化数据点
			data[idx] = make([]float64, 2)
			
			// 围绕中心点生成数据，加入一些噪声
			data[idx][0] = centers[i][0] + rand.Float64()*noise - noise/2
			data[idx][1] = centers[i][1] + rand.Float64()*noise - noise/2
		}
	}
	
	return data
}

// RunKMeansExample 演示如何使用K-means聚类算法
func RunKMeansExample(ctx context.Context) {
	// 生成示例数据：3个聚类，每个聚类50个点，噪声水平为2.0
	data := ExampleData(3, 50, 2.0)
	
	hlog.CtxInfof(ctx, "生成了3个聚类的示例数据，共计%d个数据点", len(data))
	
	// 创建K-means聚类器实例
	clusterer := NewKMeansClusterer(ClusterParams{
		K:             3,
		MaxIterations: 100,
		DistanceFunc:  EuclideanDistance,
		Tolerance:     1e-4,
	})
	
	// 执行聚类
	err := clusterer.Fit(ctx, data)
	if err != nil {
		hlog.CtxErrorf(ctx, "K-means聚类失败: %v", err)
		return
	}
	
	// 获取聚类结果
	clusters, err := clusterer.GetClusters(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "获取聚类结果失败: %v", err)
		return
	}
	
	// 获取中心点
	centroids, err := clusterer.GetCentroids(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "获取中心点失败: %v", err)
		return
	}
	
	// 输出结果
	for i, cluster := range clusters {
		hlog.CtxInfof(ctx, "聚类 #%d 包含 %d 个点，中心点位置: (%.2f, %.2f)",
			i+1, len(cluster), centroids[i][0], centroids[i][1])
	}
	
	// 测试预测功能
	newPoint := []float64{15.0, 15.0}
	clusterID, err := clusterer.Predict(ctx, newPoint)
	if err != nil {
		hlog.CtxErrorf(ctx, "预测失败: %v", err)
		return
	}
	
	hlog.CtxInfof(ctx, "新点 (%.2f, %.2f) 被预测为属于聚类 #%d",
		newPoint[0], newPoint[1], clusterID+1)
}

// RunDBSCANExample 演示如何使用DBSCAN聚类算法
func RunDBSCANExample(ctx context.Context) {
	// 生成示例数据：3个聚类，每个聚类50个点，噪声水平为3.0（较大噪声更适合DBSCAN测试）
	data := ExampleData(3, 50, 3.0)
	
	hlog.CtxInfof(ctx, "生成了3个聚类的示例数据，共计%d个数据点", len(data))
	
	// 创建DBSCAN聚类器实例
	clusterer := NewDBSCANClusterer(2.0, 5, EuclideanDistance)
	
	// 执行聚类
	err := clusterer.Fit(ctx, data)
	if err != nil {
		hlog.CtxErrorf(ctx, "DBSCAN聚类失败: %v", err)
		return
	}
	
	// 获取聚类结果
	clusters, err := clusterer.GetClusters(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "获取聚类结果失败: %v", err)
		return
	}
	
	// 输出结果
	for i, cluster := range clusters {
		hlog.CtxInfof(ctx, "聚类 #%d 包含 %d 个点", i+1, len(cluster))
	}
}
