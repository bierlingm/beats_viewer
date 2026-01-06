package chain

import (
	"fmt"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

type Store struct {
	chains    []model.Chain
	beatIndex map[string][]string // beatID -> chainIDs
}

func NewStore() *Store {
	return &Store{
		chains:    []model.Chain{},
		beatIndex: make(map[string][]string),
	}
}

func (s *Store) LoadFromCache(chains []model.Chain) {
	s.chains = chains
	s.rebuildIndex()
}

func (s *Store) rebuildIndex() {
	s.beatIndex = make(map[string][]string)
	for _, chain := range s.chains {
		for _, beatID := range chain.BeatIDs {
			s.beatIndex[beatID] = append(s.beatIndex[beatID], chain.ID)
		}
	}
}

func (s *Store) Create(name string, beatIDs []string) (*model.Chain, error) {
	if name == "" {
		return nil, fmt.Errorf("chain name required")
	}

	chain := model.Chain{
		ID:        fmt.Sprintf("chain-%d", time.Now().UnixNano()),
		Name:      name,
		BeatIDs:   beatIDs,
		CreatedAt: time.Now(),
	}

	s.chains = append(s.chains, chain)

	for _, beatID := range beatIDs {
		s.beatIndex[beatID] = append(s.beatIndex[beatID], chain.ID)
	}

	return &chain, nil
}

func (s *Store) Get(chainID string) *model.Chain {
	for i := range s.chains {
		if s.chains[i].ID == chainID {
			return &s.chains[i]
		}
	}
	return nil
}

func (s *Store) GetByName(name string) *model.Chain {
	for i := range s.chains {
		if s.chains[i].Name == name {
			return &s.chains[i]
		}
	}
	return nil
}

func (s *Store) List() []model.Chain {
	return s.chains
}

func (s *Store) AddBeat(chainID, beatID string) error {
	chain := s.Get(chainID)
	if chain == nil {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	for _, id := range chain.BeatIDs {
		if id == beatID {
			return nil // already in chain
		}
	}

	chain.BeatIDs = append(chain.BeatIDs, beatID)
	s.beatIndex[beatID] = append(s.beatIndex[beatID], chainID)

	return nil
}

func (s *Store) RemoveBeat(chainID, beatID string) error {
	chain := s.Get(chainID)
	if chain == nil {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	var newBeatIDs []string
	found := false
	for _, id := range chain.BeatIDs {
		if id == beatID {
			found = true
		} else {
			newBeatIDs = append(newBeatIDs, id)
		}
	}

	if !found {
		return fmt.Errorf("beat not in chain")
	}

	chain.BeatIDs = newBeatIDs

	var newChainIDs []string
	for _, id := range s.beatIndex[beatID] {
		if id != chainID {
			newChainIDs = append(newChainIDs, id)
		}
	}
	s.beatIndex[beatID] = newChainIDs

	return nil
}

func (s *Store) Rename(chainID, newName string) error {
	chain := s.Get(chainID)
	if chain == nil {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	chain.Name = newName
	return nil
}

func (s *Store) Delete(chainID string) error {
	var newChains []model.Chain
	var deletedChain *model.Chain

	for i := range s.chains {
		if s.chains[i].ID == chainID {
			deletedChain = &s.chains[i]
		} else {
			newChains = append(newChains, s.chains[i])
		}
	}

	if deletedChain == nil {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	for _, beatID := range deletedChain.BeatIDs {
		var newChainIDs []string
		for _, id := range s.beatIndex[beatID] {
			if id != chainID {
				newChainIDs = append(newChainIDs, id)
			}
		}
		s.beatIndex[beatID] = newChainIDs
	}

	s.chains = newChains
	return nil
}

func (s *Store) GetChainsForBeat(beatID string) []model.Chain {
	chainIDs := s.beatIndex[beatID]
	var result []model.Chain

	for _, chainID := range chainIDs {
		if chain := s.Get(chainID); chain != nil {
			result = append(result, *chain)
		}
	}

	return result
}

func (s *Store) GetBeatPosition(chainID, beatID string) (index int, total int) {
	chain := s.Get(chainID)
	if chain == nil {
		return -1, 0
	}

	for i, id := range chain.BeatIDs {
		if id == beatID {
			return i, len(chain.BeatIDs)
		}
	}

	return -1, len(chain.BeatIDs)
}

func (s *Store) GetAdjacentBeats(chainID, beatID string) (prev, next string) {
	chain := s.Get(chainID)
	if chain == nil {
		return "", ""
	}

	for i, id := range chain.BeatIDs {
		if id == beatID {
			if i > 0 {
				prev = chain.BeatIDs[i-1]
			}
			if i < len(chain.BeatIDs)-1 {
				next = chain.BeatIDs[i+1]
			}
			return
		}
	}

	return "", ""
}

func (s *Store) Export() []model.Chain {
	return s.chains
}
