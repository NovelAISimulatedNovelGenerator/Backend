package cluster

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// KMeansClusterer 实现了K-means聚类算法
type KMeansClusterer struct {
	// 聚类参数
	params ClusterParams
	
	// 聚类中心点
	centroids [][]float64
	
	// 数据点所属的聚类索引
	labels []int
	
	// 是否已完成训练
	fitted bool
}

// NewKMeansClusterer 创建一个新的K-means聚类实例
func NewKMeansClusterer(params ClusterParams) *KMeansClusterer {
	// 默认距离函数为欧几里得距离
	if params.DistanceFunc == nil {
		params.DistanceFunc = EuclideanDistance
	}
	
	// 默认最大迭代次数
	if params.MaxIterations <= 0 {
		params.MaxIterations = 100
	}
	
	// 默认收敛阈值
	if params.Tolerance <= 0 {
		params.Tolerance = 1e-4
	}
	
	return &KMeansClusterer{
		params: params,
		fitted: false,
	}
}

// Fit 执行K-means聚类算法
func (k *KMeansClusterer) Fit(ctx context.Context, data [][]float64) error {
	if len(data) == 0 {
		return errors.New("无法对空数据集进行聚类")
	}
	
	if len(data) < k.params.K {
		return errors.New("数据点数量少于请求的聚类数量")
	}
	
	hlog.CtxInfof(ctx, "开始K-means聚类，数据点数量: %d, 聚类数: %d", len(data), k.params.K)
	
	// 初始化聚类中心点
	err := k.initializeCentroids(ctx, data)
	if err != nil {
		return err
	}
	
	// 初始化标签
	k.labels = make([]int, len(data))
	
	// 记录上一次迭代的中心点，用于判断收敛
	prevCentroids := make([][]float64, len(k.centroids))
	for i := range k.centroids {
		prevCentroids[i] = make([]float64, len(k.centroids[i]))
	}
	
	// 主迭代过程
	for iter := 0; iter < k.params.MaxIterations; iter++ {
		// 1. 分配步骤：将每个数据点分配到最近的中心点
		for i, point := range data {
			k.labels[i] = k.findNearestCentroid(point)
		}
		
		// 2. 保存当前中心点
		for i, centroid := range k.centroids {
			copy(prevCentroids[i], centroid)
		}
		
		// 3. 更新步骤：重新计算每个聚类的中心点
		err := k.updateCentroids(data)
		if err != nil {
			return err
		}
		
		// 4. 检查收敛性
		if k.hasConverged(prevCentroids) {
			hlog.CtxInfof(ctx, "K-means算法已收敛，迭代次数: %d", iter+1)
			break
		}
	}
	
	k.fitted = true
	hlog.CtxInfof(ctx, "K-means聚类完成")
	return nil
}

// initializeCentroids 初始化聚类中心点
func (k *KMeansClusterer) initializeCentroids(ctx context.Context, data [][]float64) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// 获取数据维度
	if len(data) == 0 {
		return errors.New("数据集为空")
	}
	dim := len(data[0])
	
	// 初始化中心点
	k.centroids = make([][]float64, k.params.K)
	
	// 使用随机选择的数据点作为初始中心点
	indices := rng.Perm(len(data))
	for i := 0; i < k.params.K; i++ {
		k.centroids[i] = make([]float64, dim)
		copy(k.centroids[i], data[indices[i]])
	}
	
	return nil
}

// findNearestCentroid 查找最近的聚类中心点
func (k *KMeansClusterer) findNearestCentroid(point []float64) int {
	minDist := math.Inf(1)
	var nearestIdx int
	
	for i, centroid := range k.centroids {
		dist := k.params.DistanceFunc(point, centroid)
		if dist < minDist {
			minDist = dist
			nearestIdx = i
		}
	}
	
	return nearestIdx
}

// updateCentroids 更新聚类中心点
func (k *KMeansClusterer) updateCentroids(data [][]float64) error {
	// 获取数据维度
	if len(data) == 0 {
		return errors.New("数据集为空")
	}
	dim := len(data[0])
	
	// 初始化新中心点和计数器
	newCentroids := make([][]float64, k.params.K)
	counts := make([]int, k.params.K)
	
	for i := range newCentroids {
		newCentroids[i] = make([]float64, dim)
	}
	
	// 累加每个聚类中的所有点
	for i, label := range k.labels {
		counts[label]++
		for j := 0; j < dim; j++ {
			newCentroids[label][j] += data[i][j]
		}
	}
	
	// 计算平均值作为新的中心点
	for i := 0; i < k.params.K; i++ {
		if counts[i] == 0 {
			// 处理空聚类
			continue
		}
		
		for j := 0; j < dim; j++ {
			newCentroids[i][j] /= float64(counts[i])
		}
	}
	
	// 更新中心点
	k.centroids = newCentroids
	return nil
}

// hasConverged 判断算法是否已收敛
func (k *KMeansClusterer) hasConverged(prevCentroids [][]float64) bool {
	for i, centroid := range k.centroids {
		dist := k.params.DistanceFunc(centroid, prevCentroids[i])
		if dist > k.params.Tolerance {
			return false
		}
	}
	return true
}

// Predict 预测新数据点所属的聚类
func (k *KMeansClusterer) Predict(ctx context.Context, point []float64) (int, error) {
	if !k.fitted {
		return -1, errors.New("模型尚未训练")
	}
	
	return k.findNearestCentroid(point), nil
}

// GetClusters 获取所有聚类及其包含的数据点索引
func (k *KMeansClusterer) GetClusters(ctx context.Context) ([][]int, error) {
	if !k.fitted {
		return nil, errors.New("模型尚未训练")
	}
	
	// 初始化聚类列表
	clusters := make([][]int, k.params.K)
	for i := range clusters {
		clusters[i] = make([]int, 0)
	}
	
	// 将数据点分配到各个聚类
	for i, label := range k.labels {
		clusters[label] = append(clusters[label], i)
	}
	
	return clusters, nil
}

// GetCentroids 获取所有聚类的中心点
func (k *KMeansClusterer) GetCentroids(ctx context.Context) ([][]float64, error) {
	if !k.fitted {
		return nil, errors.New("模型尚未训练")
	}
	
	// 创建中心点的副本
	centroids := make([][]float64, len(k.centroids))
	for i, c := range k.centroids {
		centroids[i] = make([]float64, len(c))
		copy(centroids[i], c)
	}
	
	return centroids, nil
}
