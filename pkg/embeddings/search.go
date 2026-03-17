package embeddings

import (
	"context"
	"math"
	"sort"

	"github.com/bierlingm/beats_viewer/pkg/cluster"
	"github.com/bierlingm/beats_viewer/pkg/model"
)

// SemanticResult represents a search result with similarity score
type SemanticResult struct {
	BeatID  string  `json:"beat_id"`
	Score   float64 `json:"score"`
	Preview string  `json:"preview"`
}

// ComputeResult tracks embedding generation progress
type ComputeResult struct {
	Computed int      `json:"computed"`
	Skipped  int      `json:"skipped"`
	Errors   int      `json:"errors"`
	ErrorIDs []string `json:"error_ids,omitempty"`
}

// ComputeMissing generates embeddings for beats that don't have them
func ComputeMissing(ctx context.Context, beats []model.Beat, store *Store, ollama *cluster.OllamaClient) ComputeResult {
	result := ComputeResult{}

	for _, beat := range beats {
		if store.Has(beat.ID) {
			result.Skipped++
			continue
		}

		embedding, err := ollama.GetEmbedding(ctx, beat.Content)
		if err != nil {
			result.Errors++
			result.ErrorIDs = append(result.ErrorIDs, beat.ID)
			continue
		}

		if err := store.Put(beat.ID, embedding); err != nil {
			result.Errors++
			result.ErrorIDs = append(result.ErrorIDs, beat.ID)
			continue
		}

		result.Computed++
	}

	return result
}

// SemanticSearch finds beats similar to query text
func SemanticSearch(ctx context.Context, query string, beats []model.Beat, store *Store, ollama *cluster.OllamaClient, limit int) ([]SemanticResult, error) {
	queryEmb, err := ollama.GetEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []SemanticResult

	for _, beat := range beats {
		beatEmb, err := store.Get(beat.ID)
		if err != nil {
			continue // Skip beats without embeddings
		}

		score := cosineSimilarity(queryEmb, beatEmb)

		results = append(results, SemanticResult{
			BeatID:  beat.ID,
			Score:   score,
			Preview: beat.ContentPreview(100),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FindSimilar finds beats similar to a given beat
func FindSimilar(beatID string, beats []model.Beat, store *Store, limit int) ([]SemanticResult, error) {
	targetEmb, err := store.Get(beatID)
	if err != nil {
		return nil, err
	}

	var results []SemanticResult

	for _, beat := range beats {
		if beat.ID == beatID {
			continue
		}

		beatEmb, err := store.Get(beat.ID)
		if err != nil {
			continue
		}

		score := cosineSimilarity(targetEmb, beatEmb)

		results = append(results, SemanticResult{
			BeatID:  beat.ID,
			Score:   score,
			Preview: beat.ContentPreview(100),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
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
