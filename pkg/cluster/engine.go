package cluster

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

const (
	DefaultK        = 8
	MaxIterations   = 100
	MinClusterSize  = 2
)

type Engine struct {
	ollama         *OllamaClient
	embeddingCache map[string][]float64
}

func NewEngine() *Engine {
	return &Engine{
		ollama:         NewOllamaClient(),
		embeddingCache: make(map[string][]float64),
	}
}

func (e *Engine) IsAvailable() bool {
	return e.ollama.IsAvailable()
}

func (e *Engine) Refresh() bool {
	return e.ollama.Refresh()
}

func (e *Engine) SetEmbeddingCache(cache map[string][]float64) {
	if cache != nil {
		e.embeddingCache = cache
	}
}

func (e *Engine) GenerateClusters(ctx context.Context, beats []model.EnrichedBeat, k int) ([]model.Cluster, error) {
	if !e.ollama.IsAvailable() {
		return nil, fmt.Errorf("ollama not available")
	}

	if k <= 0 {
		k = DefaultK
	}

	embeddings := make([][]float64, 0, len(beats))
	beatIndices := make([]int, 0, len(beats))

	for i, beat := range beats {
		emb, err := e.getEmbedding(ctx, beat)
		if err != nil {
			continue
		}
		embeddings = append(embeddings, emb)
		beatIndices = append(beatIndices, i)
	}

	if len(embeddings) < k {
		k = len(embeddings)
	}

	if k < 2 {
		return nil, fmt.Errorf("not enough beats for clustering")
	}

	assignments, centroids := KMeans(embeddings, k, MaxIterations)

	clusterBeats := make(map[int][]int)
	for i, cluster := range assignments {
		clusterBeats[cluster] = append(clusterBeats[cluster], beatIndices[i])
	}

	var clusters []model.Cluster
	for clusterIdx, beatIdxs := range clusterBeats {
		if len(beatIdxs) < MinClusterSize {
			continue
		}

		var beatIDs []string
		var contents []string
		var totalRipeness float64

		for _, idx := range beatIdxs {
			beatIDs = append(beatIDs, beats[idx].ID)
			contents = append(contents, beats[idx].Content)
			totalRipeness += beats[idx].RipenessScore
		}

		avgRipeness := totalRipeness / float64(len(beatIdxs))

		cluster := model.Cluster{
			ID:            fmt.Sprintf("cluster-%d-%d", time.Now().Unix(), clusterIdx),
			Name:          generateClusterName(contents),
			BeatIDs:       beatIDs,
			Centroid:      centroids[clusterIdx],
			Keywords:      extractKeywords(contents),
			CreatedAt:     time.Now(),
			RipenessScore: avgRipeness,
		}

		clusters = append(clusters, cluster)
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].RipenessScore > clusters[j].RipenessScore
	})

	return clusters, nil
}

func (e *Engine) getEmbedding(ctx context.Context, beat model.EnrichedBeat) ([]float64, error) {
	if emb, ok := e.embeddingCache[beat.ID]; ok {
		return emb, nil
	}

	emb, err := e.ollama.GetEmbedding(ctx, beat.Content)
	if err != nil {
		return nil, err
	}

	e.embeddingCache[beat.ID] = emb
	return emb, nil
}

func (e *Engine) FindSimilar(ctx context.Context, beat model.EnrichedBeat, allBeats []model.EnrichedBeat, limit int) ([]model.EnrichedBeat, error) {
	if !e.ollama.IsAvailable() {
		return nil, fmt.Errorf("ollama not available")
	}

	targetEmb, err := e.getEmbedding(ctx, beat)
	if err != nil {
		return nil, fmt.Errorf("getting target embedding: %w", err)
	}

	type scored struct {
		beat  model.EnrichedBeat
		score float64
	}

	var scored_beats []scored
	for _, other := range allBeats {
		if other.ID == beat.ID {
			continue
		}

		otherEmb, err := e.getEmbedding(ctx, other)
		if err != nil {
			continue
		}

		similarity := CosineSimilarity(targetEmb, otherEmb)
		scored_beats = append(scored_beats, scored{other, similarity})
	}

	sort.Slice(scored_beats, func(i, j int) bool {
		return scored_beats[i].score > scored_beats[j].score
	})

	if limit > len(scored_beats) {
		limit = len(scored_beats)
	}

	var result []model.EnrichedBeat
	for i := 0; i < limit; i++ {
		result = append(result, scored_beats[i].beat)
	}

	return result, nil
}

func (e *Engine) GetEmbeddingCache() map[string][]float64 {
	return e.embeddingCache
}

func generateClusterName(contents []string) string {
	words := extractKeywords(contents)
	if len(words) == 0 {
		return "Unnamed Cluster"
	}

	if len(words) > 3 {
		words = words[:3]
	}

	return strings.Join(words, " & ")
}

func extractKeywords(contents []string) []string {
	wordFreq := make(map[string]int)
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
		"we": true, "they": true, "what": true, "which": true, "who": true,
		"when": true, "where": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "not": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true,
		"can": true, "about": true, "into": true, "through": true, "during": true,
		"before": true, "after": true, "above": true, "below": true, "up": true,
		"down": true, "out": true, "off": true, "over": true, "under": true,
		"again": true, "further": true, "then": true, "once": true,
	}

	for _, content := range contents {
		words := strings.Fields(strings.ToLower(content))
		for _, word := range words {
			word = strings.Trim(word, ".,!?\"'()[]{}")
			if len(word) < 3 || stopWords[word] {
				continue
			}
			wordFreq[word]++
		}
	}

	type wordCount struct {
		word  string
		count int
	}
	var counts []wordCount
	for w, c := range wordFreq {
		counts = append(counts, wordCount{w, c})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	var keywords []string
	for i := 0; i < 5 && i < len(counts); i++ {
		keywords = append(keywords, counts[i].word)
	}

	return keywords
}
