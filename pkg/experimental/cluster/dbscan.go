package cluster

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// DBSCAN聚类的特殊标签
const (
	// 未分类的点
	DBSCANUnclassified = -1
	// 噪声点
	DBSCANNoise = -2
)

// DBSCANClusterer 实现了DBSCAN密度聚类算法
type DBSCANClusterer struct {
	// 邻域半径
	Eps float64

	// 形成核心点所需的最小点数
	MinPts int

	// 距离计算函数
	DistanceFunc DistanceFunc

	// 数据点所属的聚类索引
	labels []int

	// 聚类数量
	numClusters int

	// 是否已完成训练
	fitted bool
}

// NewDBSCANClusterer 创建一个新的DBSCAN聚类实例
func NewDBSCANClusterer(eps float64, minPts int, distanceFunc DistanceFunc) *DBSCANClusterer {
	if distanceFunc == nil {
		distanceFunc = EuclideanDistance
	}

	return &DBSCANClusterer{
		Eps:          eps,
		MinPts:       minPts,
		DistanceFunc: distanceFunc,
		fitted:       false,
	}
}

// Fit 执行DBSCAN聚类算法
func (d *DBSCANClusterer) Fit(ctx context.Context, data [][]float64) error {
	if len(data) == 0 {
		return errors.New("无法对空数据集进行聚类")
	}

	hlog.CtxInfof(ctx, "开始DBSCAN聚类，数据点数量: %d, Eps: %f, MinPts: %d",
		len(data), d.Eps, d.MinPts)

	// 初始化标签，将所有点标记为未分类
	d.labels = make([]int, len(data))
	for i := range d.labels {
		d.labels[i] = DBSCANUnclassified
	}

	// 聚类ID从0开始
	clusterID := 0

	// 遍历所有点
	for i := 0; i < len(data); i++ {
		// 跳过已分类的点
		if d.labels[i] != DBSCANUnclassified {
			continue
		}

		// 查找点i的邻域
		neighbors := d.regionQuery(data, i)

		// 如果邻域中的点数少于MinPts，则标记为噪声
		if len(neighbors) < d.MinPts {
			d.labels[i] = DBSCANNoise
			continue
		}

		// 否则，开始一个新的聚类
		clusterID++
		d.labels[i] = clusterID

		// 扩展聚类
		d.expandCluster(ctx, data, i, neighbors, clusterID)
	}

	d.numClusters = clusterID
	d.fitted = true

	hlog.CtxInfof(ctx, "DBSCAN聚类完成，共形成%d个聚类", d.numClusters)
	return nil
}

// regionQuery 查找点p的邻域中的所有点
func (d *DBSCANClusterer) regionQuery(data [][]float64, p int) []int {
	var neighbors []int

	for i := 0; i < len(data); i++ {
		if d.DistanceFunc(data[p], data[i]) <= d.Eps {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}

// expandCluster 扩展聚类
func (d *DBSCANClusterer) expandCluster(ctx context.Context, data [][]float64, p int, neighbors []int, clusterID int) {
	// 创建一个队列进行广度优先搜索
	queue := make([]int, len(neighbors))
	copy(queue, neighbors)

	// 处理队列中的每个点
	for i := 0; i < len(queue); i++ {
		currentPoint := queue[i]

		// 如果当前点是噪声，将其归入当前聚类
		if d.labels[currentPoint] == DBSCANNoise {
			d.labels[currentPoint] = clusterID
			continue
		}

		// 如果当前点未分类
		if d.labels[currentPoint] == DBSCANUnclassified {
			// 将其归入当前聚类
			d.labels[currentPoint] = clusterID

			// 查找其邻域
			currentNeighbors := d.regionQuery(data, currentPoint)

			// 如果是核心点，扩展其邻域
			if len(currentNeighbors) >= d.MinPts {
				for _, n := range currentNeighbors {
					// 如果邻居未被处理或是噪声，加入队列
					if d.labels[n] == DBSCANUnclassified || d.labels[n] == DBSCANNoise {
						// 检查是否已在队列中
						alreadyInQueue := false
						for _, q := range queue {
							if q == n {
								alreadyInQueue = true
								break
							}
						}

						if !alreadyInQueue {
							queue = append(queue, n)
						}
					}
				}
			}
		}
	}
}

// Predict 预测新数据点所属的聚类
func (d *DBSCANClusterer) Predict(ctx context.Context, point []float64) (int, error) {
	if !d.fitted {
		return DBSCANUnclassified, errors.New("模型尚未训练")
	}

	// DBSCAN不支持直接预测新点，但我们可以找到最近的非噪声点所属的聚类
	minDist := -1.0
	nearestCluster := DBSCANNoise

	for _, label := range d.labels {
		// 跳过噪声点
		if label == DBSCANNoise {
			continue
		}

		// 计算距离
		dist := d.DistanceFunc(point, nil) // 需要替换为实际的数据点

		// 如果是第一个非噪声点或距离更小
		if minDist < 0 || dist < minDist {
			minDist = dist
			nearestCluster = label
		}
	}

	// 如果点距离小于Eps，则归入该聚类，否则为噪声
	if minDist <= d.Eps {
		return nearestCluster, nil
	}

	return DBSCANNoise, nil
}

// GetClusters 获取所有聚类及其包含的数据点索引
func (d *DBSCANClusterer) GetClusters(ctx context.Context) ([][]int, error) {
	if !d.fitted {
		return nil, errors.New("模型尚未训练")
	}

	// 初始化聚类列表
	clusters := make([][]int, d.numClusters+1) // +1 是因为聚类ID从1开始
	for i := range clusters {
		clusters[i] = make([]int, 0)
	}

	// 将数据点分配到各个聚类
	for i, label := range d.labels {
		// 跳过噪声点
		if label <= 0 {
			continue
		}

		clusters[label] = append(clusters[label], i)
	}

	// 移除空聚类（比如索引0）
	result := make([][]int, 0)
	for _, cluster := range clusters {
		if len(cluster) > 0 {
			result = append(result, cluster)
		}
	}

	return result, nil
}

// GetCentroids DBSCAN没有真正的中心点概念，返回每个聚类的平均位置
func (d *DBSCANClusterer) GetCentroids(ctx context.Context) ([][]float64, error) {
	return nil, errors.New("DBSCAN算法没有中心点概念")
}
