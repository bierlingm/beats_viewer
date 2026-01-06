package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"beats_viewer/pkg/model"
)

// LoadCache reads the cache file from the beats directory
func LoadCache(beatsDir string) (*model.Cache, error) {
	cachePath := filepath.Join(beatsDir, model.CacheFileName)

	file, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening cache: %w", err)
	}
	defer file.Close()

	var cache model.Cache
	if err := json.NewDecoder(file).Decode(&cache); err != nil {
		return nil, fmt.Errorf("decoding cache: %w", err)
	}

	return &cache, nil
}

// SaveCache writes the cache to the beats directory atomically
func SaveCache(beatsDir string, cache *model.Cache) error {
	cachePath := filepath.Join(beatsDir, model.CacheFileName)
	tmpPath := cachePath + ".tmp"

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("writing temp cache: %w", err)
	}

	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming cache: %w", err)
	}

	return nil
}

// ComputeSourceHash calculates a hash of beats.jsonl for cache invalidation
func ComputeSourceHash(beatsDir string) (string, error) {
	filePath := filepath.Join(beatsDir, BeatsFile)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("opening beats file: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("hashing beats file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// IsCacheValid checks if the cache is valid for the current beats data
func IsCacheValid(beatsDir string, cache *model.Cache) bool {
	if cache == nil {
		return false
	}

	if cache.Version != model.CacheVersion {
		return false
	}

	currentHash, err := ComputeSourceHash(beatsDir)
	if err != nil {
		return false
	}

	return cache.SourceHash == currentHash
}

// LoadOrCreateCache loads existing cache or returns nil if rebuild needed
func LoadOrCreateCache(beatsDir string) (*model.Cache, bool, error) {
	cache, err := LoadCache(beatsDir)
	if err != nil {
		return nil, false, err
	}

	if IsCacheValid(beatsDir, cache) {
		return cache, false, nil
	}

	return nil, true, nil
}
