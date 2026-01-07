package inference

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	BeadsDirName  = ".beads"
	IssuesFile    = "issues.jsonl"
)

// Bead represents a minimal bead (issue) from the beads tracker
type Bead struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

// BeadsStore provides read-only access to beads data
type BeadsStore struct {
	Dir   string
	Beads []Bead
}

// LoadBeads reads beads from .beads/issues.jsonl
func LoadBeads(beadsDir string) (*BeadsStore, error) {
	issuesPath := filepath.Join(beadsDir, IssuesFile)

	file, err := os.Open(issuesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &BeadsStore{Dir: beadsDir, Beads: []Bead{}}, nil
		}
		return nil, fmt.Errorf("opening issues file: %w", err)
	}
	defer file.Close()

	var beads []Bead
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var bead Bead
		if err := json.Unmarshal(line, &bead); err != nil {
			// Skip malformed lines but continue processing
			continue
		}
		beads = append(beads, bead)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading issues file: %w", err)
	}

	return &BeadsStore{Dir: beadsDir, Beads: beads}, nil
}

// FindBeadsDir walks up from startDir looking for .beads directory
func FindBeadsDir(startDir string) (string, error) {
	dir := startDir

	for {
		beadsPath := filepath.Join(dir, BeadsDirName)
		if info, err := os.Stat(beadsPath); err == nil && info.IsDir() {
			return beadsPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "", fmt.Errorf("no %s directory found", BeadsDirName)
		}
		dir = parent
	}
}

// LoadBeadsFromBeatsDir finds and loads beads relative to a beats directory
func LoadBeadsFromBeatsDir(beatsDir string) (*BeadsStore, error) {
	// Start from parent of .beats directory
	startDir := filepath.Dir(beatsDir)

	beadsDir, err := FindBeadsDir(startDir)
	if err != nil {
		// No beads directory found - return nil store (graceful)
		return nil, nil
	}

	return LoadBeads(beadsDir)
}

// GetBead returns a bead by ID
func (s *BeadsStore) GetBead(id string) *Bead {
	if s == nil {
		return nil
	}
	for i := range s.Beads {
		if s.Beads[i].ID == id {
			return &s.Beads[i]
		}
	}
	return nil
}

// OpenBeads returns beads that are not closed
func (s *BeadsStore) OpenBeads() []Bead {
	if s == nil {
		return nil
	}
	var result []Bead
	for _, b := range s.Beads {
		if b.ClosedAt == nil {
			result = append(result, b)
		}
	}
	return result
}

// ClosedBeads returns beads that have been closed
func (s *BeadsStore) ClosedBeads() []Bead {
	if s == nil {
		return nil
	}
	var result []Bead
	for _, b := range s.Beads {
		if b.ClosedAt != nil {
			result = append(result, b)
		}
	}
	return result
}

// BeadsInRange returns beads created within a time range
func (s *BeadsStore) BeadsInRange(start, end time.Time) []Bead {
	if s == nil {
		return nil
	}
	var result []Bead
	for _, b := range s.Beads {
		if !b.CreatedAt.Before(start) && !b.CreatedAt.After(end) {
			result = append(result, b)
		}
	}
	return result
}

// Count returns the total number of beads
func (s *BeadsStore) Count() int {
	if s == nil {
		return 0
	}
	return len(s.Beads)
}

// HasBeads returns true if store has any beads
func (s *BeadsStore) HasBeads() bool {
	return s != nil && len(s.Beads) > 0
}
