package cluster

import (
	"math"
	"math/rand"
)

func KMeans(embeddings [][]float64, k int, maxIter int) ([]int, [][]float64) {
	if len(embeddings) == 0 || k <= 0 {
		return nil, nil
	}

	if k > len(embeddings) {
		k = len(embeddings)
	}

	dim := len(embeddings[0])
	centroids := initCentroids(embeddings, k)
	assignments := make([]int, len(embeddings))

	for iter := 0; iter < maxIter; iter++ {
		changed := false

		for i, emb := range embeddings {
			nearest := findNearest(emb, centroids)
			if assignments[i] != nearest {
				assignments[i] = nearest
				changed = true
			}
		}

		if !changed {
			break
		}

		centroids = updateCentroids(embeddings, assignments, k, dim)
	}

	return assignments, centroids
}

func initCentroids(embeddings [][]float64, k int) [][]float64 {
	n := len(embeddings)
	if n == 0 {
		return nil
	}

	dim := len(embeddings[0])
	centroids := make([][]float64, k)

	used := make(map[int]bool)
	for i := 0; i < k; i++ {
		idx := rand.Intn(n)
		for used[idx] && len(used) < n {
			idx = rand.Intn(n)
		}
		used[idx] = true

		centroids[i] = make([]float64, dim)
		copy(centroids[i], embeddings[idx])
	}

	return centroids
}

func findNearest(point []float64, centroids [][]float64) int {
	minDist := math.MaxFloat64
	nearest := 0

	for i, c := range centroids {
		dist := euclideanDistance(point, c)
		if dist < minDist {
			minDist = dist
			nearest = i
		}
	}

	return nearest
}

func updateCentroids(embeddings [][]float64, assignments []int, k, dim int) [][]float64 {
	centroids := make([][]float64, k)
	counts := make([]int, k)

	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, dim)
	}

	for i, emb := range embeddings {
		cluster := assignments[i]
		counts[cluster]++
		for j, val := range emb {
			centroids[cluster][j] += val
		}
	}

	for i := 0; i < k; i++ {
		if counts[i] > 0 {
			for j := 0; j < dim; j++ {
				centroids[i][j] /= float64(counts[i])
			}
		}
	}

	return centroids
}

func euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
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
