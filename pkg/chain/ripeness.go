package chain

import "beats_viewer/pkg/model"

func CalculateRipeness(chain *model.Chain, ripenessScores map[string]float64) float64 {
	if len(chain.BeatIDs) == 0 {
		return 0
	}

	var total float64
	var count int

	for _, beatID := range chain.BeatIDs {
		if score, ok := ripenessScores[beatID]; ok {
			total += score
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

func UpdateChainRipeness(chain *model.Chain, ripenessScores map[string]float64) {
	chain.RipenessScore = CalculateRipeness(chain, ripenessScores)
}

func UpdateAllChainRipeness(chains []model.Chain, ripenessScores map[string]float64) {
	for i := range chains {
		UpdateChainRipeness(&chains[i], ripenessScores)
	}
}
