package inference

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadBeads(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "beads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write test JSONL
	content := `{"id":"mb-001","title":"First bead","description":"Test desc","created_at":"2024-01-15T10:00:00Z"}
{"id":"mb-002","title":"Second bead","description":"Another","created_at":"2024-01-16T11:00:00Z","closed_at":"2024-01-20T15:00:00Z"}
`
	err = os.WriteFile(filepath.Join(tmpDir, "issues.jsonl"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	store, err := LoadBeads(tmpDir)
	if err != nil {
		t.Fatalf("LoadBeads failed: %v", err)
	}

	if store.Count() != 2 {
		t.Errorf("expected 2 beads, got %d", store.Count())
	}

	// Check first bead
	bead := store.GetBead("mb-001")
	if bead == nil {
		t.Fatal("expected to find mb-001")
	}
	if bead.Title != "First bead" {
		t.Errorf("expected 'First bead', got %q", bead.Title)
	}

	// Check open/closed filtering
	if len(store.OpenBeads()) != 1 {
		t.Errorf("expected 1 open bead, got %d", len(store.OpenBeads()))
	}
	if len(store.ClosedBeads()) != 1 {
		t.Errorf("expected 1 closed bead, got %d", len(store.ClosedBeads()))
	}
}

func TestLoadBeads_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write empty file
	err = os.WriteFile(filepath.Join(tmpDir, "issues.jsonl"), []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	store, err := LoadBeads(tmpDir)
	if err != nil {
		t.Fatalf("LoadBeads failed: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("expected 0 beads, got %d", store.Count())
	}
}

func TestLoadBeads_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := LoadBeads(tmpDir)
	if err != nil {
		t.Fatalf("LoadBeads should not error on missing file: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("expected 0 beads for missing file, got %d", store.Count())
	}
}

func TestLoadBeads_MalformedLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write file with one good line and one malformed
	content := `{"id":"mb-001","title":"Good bead","created_at":"2024-01-15T10:00:00Z"}
not json at all
{"id":"mb-002","title":"Another good","created_at":"2024-01-16T11:00:00Z"}
`
	err = os.WriteFile(filepath.Join(tmpDir, "issues.jsonl"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	store, err := LoadBeads(tmpDir)
	if err != nil {
		t.Fatalf("LoadBeads should skip malformed lines: %v", err)
	}

	// Should have loaded 2 valid beads
	if store.Count() != 2 {
		t.Errorf("expected 2 beads (skipping malformed), got %d", store.Count())
	}
}

func TestFindBeadsDir(t *testing.T) {
	// Create nested structure
	tmpDir, err := os.MkdirTemp("", "beads_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .beads dir and nested structure
	beadsDir := filepath.Join(tmpDir, ".beads")
	nestedDir := filepath.Join(tmpDir, "sub", "deep")
	os.MkdirAll(beadsDir, 0755)
	os.MkdirAll(nestedDir, 0755)

	// Should find .beads from nested dir
	found, err := FindBeadsDir(nestedDir)
	if err != nil {
		t.Fatalf("FindBeadsDir failed: %v", err)
	}

	if found != beadsDir {
		t.Errorf("expected %s, got %s", beadsDir, found)
	}
}

func TestBeadsInRange(t *testing.T) {
	store := &BeadsStore{
		Beads: []Bead{
			{ID: "1", CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
			{ID: "2", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{ID: "3", CreatedAt: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		},
	}

	start := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC)

	result := store.BeadsInRange(start, end)
	if len(result) != 1 {
		t.Errorf("expected 1 bead in range, got %d", len(result))
	}
	if len(result) > 0 && result[0].ID != "2" {
		t.Errorf("expected bead 2, got %s", result[0].ID)
	}
}

func TestNilBeadsStore(t *testing.T) {
	var store *BeadsStore

	// All methods should handle nil gracefully
	if store.GetBead("x") != nil {
		t.Error("GetBead on nil should return nil")
	}
	if store.OpenBeads() != nil {
		t.Error("OpenBeads on nil should return nil")
	}
	if store.ClosedBeads() != nil {
		t.Error("ClosedBeads on nil should return nil")
	}
	if store.Count() != 0 {
		t.Error("Count on nil should return 0")
	}
	if store.HasBeads() {
		t.Error("HasBeads on nil should return false")
	}
}
