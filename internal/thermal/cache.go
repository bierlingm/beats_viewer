package thermal

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

const cacheFileName = "temperature.json"

// SaveTemperatureCache writes the temperature output to .wald/temperature.json.
func SaveTemperatureCache(werkRoot string, output *TemperatureOutput) error {
	waldDir := filepath.Join(werkRoot, ".wald")
	if err := os.MkdirAll(waldDir, 0755); err != nil {
		return err
	}

	cachePath := filepath.Join(waldDir, cacheFileName)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// LoadTemperatureCache reads the previous temperature output from .wald/temperature.json.
func LoadTemperatureCache(werkRoot string) (*TemperatureOutput, error) {
	cachePath := filepath.Join(werkRoot, ".wald", cacheFileName)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var output TemperatureOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

// ComputeTrends updates current output with trends based on previous cache.
func ComputeTrends(current *TemperatureOutput, previous *TemperatureOutput, config *Config) {
	if config == nil {
		config = DefaultConfig()
	}

	for path, dirTemp := range current.Directories {
		if previous != nil {
			if prevDir, ok := previous.Directories[path]; ok {
				trend, delta := ComputeTrend(dirTemp.Temperature, prevDir.Temperature, config)
				dirTemp.Trend = trend
				dirTemp.TrendDelta = delta
				continue
			}
		}
		// No previous data - mark as new/stable with zero delta
		dirTemp.Trend = "stable"
		dirTemp.TrendDelta = 0
	}

	// Same for cooperators if present
	for path, coopTemp := range current.Cooperators {
		if previous != nil {
			if prevCoop, ok := previous.Cooperators[path]; ok {
				trend, delta := ComputeTrend(coopTemp.Temperature, prevCoop.Temperature, config)
				coopTemp.Trend = trend
				coopTemp.TrendDelta = delta
				continue
			}
		}
		coopTemp.Trend = "stable"
		coopTemp.TrendDelta = 0
	}
}

// ComputeTemperatureWithCache computes temperatures, loads previous cache for trends, and saves new cache.
func ComputeTemperatureWithCache(beats []Beat, clusters []model.Cluster, directories []WALDDirectory, config *Config, werkRoot string) (*TemperatureOutput, error) {
	// Compute current temperatures
	output := ComputeTemperature(beats, clusters, directories, config, werkRoot)

	// Load previous cache for trend computation
	previous, _ := LoadTemperatureCache(werkRoot)
	// Ignore error - missing cache is fine for first run

	// Compute trends
	ComputeTrends(output, previous, config)

	// Save updated cache
	if err := SaveTemperatureCache(werkRoot, output); err != nil {
		return output, err
	}

	return output, nil
}
