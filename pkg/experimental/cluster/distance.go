package cluster

import (
	"math"
)

// EuclideanDistance 计算两个向量之间的欧几里得距离
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1) // 返回正无穷大表示错误
	}
	
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	
	return math.Sqrt(sum)
}

// ManhattanDistance 计算两个向量之间的曼哈顿距离
func ManhattanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1) // 返回正无穷大表示错误
	}
	
	var sum float64
	for i := range a {
		sum += math.Abs(a[i] - b[i])
	}
	
	return sum
}

// CosineSimilarity 计算两个向量之间的余弦相似度
// 注意：余弦相似度越大表示越相似，与距离相反
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return -1 // 返回-1表示错误
	}
	
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// CosineDistance 将余弦相似度转换为距离度量
func CosineDistance(a, b []float64) float64 {
	similarity := CosineSimilarity(a, b)
	// 转换为距离：1 - 相似度
	return 1 - similarity
}
