package embeddings

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	EmbeddingsFile = "embeddings.bin"
	IndexFile      = "embeddings.idx"
	EmbeddingDim   = 768 // nomic-embed-text dimension
)

// Store manages beat embeddings in binary format
type Store struct {
	dir   string
	index map[string]int64 // beatID -> offset in bin file
	mu    sync.RWMutex
}

// NewStore creates or opens an embedding store
func NewStore(beatsDir string) (*Store, error) {
	store := &Store{
		dir:   beatsDir,
		index: make(map[string]int64),
	}

	if err := store.loadIndex(); err != nil {
		// Index doesn't exist yet, that's OK
	}

	return store, nil
}

func (s *Store) loadIndex() error {
	path := filepath.Join(s.dir, IndexFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			offset, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				s.index[parts[0]] = offset
			}
		}
	}

	return nil
}

func (s *Store) saveIndex() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var lines []string
	for id, offset := range s.index {
		lines = append(lines, fmt.Sprintf("%s\t%d", id, offset))
	}

	path := filepath.Join(s.dir, IndexFile)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// Has checks if a beat has an embedding
func (s *Store) Has(beatID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.index[beatID]
	return ok
}

// Put stores an embedding for a beat
func (s *Store) Put(beatID string, embedding []float64) error {
	if len(embedding) != EmbeddingDim {
		return fmt.Errorf("expected %d dimensions, got %d", EmbeddingDim, len(embedding))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, EmbeddingsFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	offset := info.Size()

	for _, v := range embedding {
		if err := binary.Write(f, binary.LittleEndian, v); err != nil {
			return err
		}
	}

	s.index[beatID] = offset

	// Save index inline
	var lines []string
	for id, off := range s.index {
		lines = append(lines, fmt.Sprintf("%s\t%d", id, off))
	}
	idxPath := filepath.Join(s.dir, IndexFile)
	return os.WriteFile(idxPath, []byte(strings.Join(lines, "\n")), 0644)
}

// Get retrieves an embedding for a beat
func (s *Store) Get(beatID string) ([]float64, error) {
	s.mu.RLock()
	offset, ok := s.index[beatID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("embedding not found for %s", beatID)
	}

	path := filepath.Join(s.dir, EmbeddingsFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, 0); err != nil {
		return nil, err
	}

	embedding := make([]float64, EmbeddingDim)
	for i := range embedding {
		if err := binary.Read(f, binary.LittleEndian, &embedding[i]); err != nil {
			return nil, err
		}
	}

	return embedding, nil
}

// Coverage returns stored count
func (s *Store) Coverage() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.index)
}

// IDs returns all beat IDs with embeddings
func (s *Store) IDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.index))
	for id := range s.index {
		ids = append(ids, id)
	}
	return ids
}
