package loader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

const (
	BeatsDir  = ".beats"
	BeatsFile = "beats.jsonl"
)

func FindBeatsDir(startPath string) (string, error) {
	current := startPath
	for {
		candidate := filepath.Join(current, BeatsDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no %s directory found", BeatsDir)
		}
		current = parent
	}
}

func LoadBeats(beatsDir string) ([]model.Beat, error) {
	filePath := filepath.Join(beatsDir, BeatsFile)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Beat{}, nil
		}
		return nil, fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer file.Close()

	var beats []model.Beat
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var beat model.Beat
		if err := json.Unmarshal([]byte(line), &beat); err != nil {
			continue
		}
		beats = append(beats, beat)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	sort.Slice(beats, func(i, j int) bool {
		return beats[i].CreatedAt.After(beats[j].CreatedAt)
	})

	return beats, nil
}

func DiscoverProjects(rootPath string) ([]model.Project, error) {
	var projects []model.Project

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == "vendor" {
			return filepath.SkipDir
		}
		if info.Name() == BeatsDir {
			parentDir := filepath.Dir(path)
			projectName := filepath.Base(parentDir)
			if parentDir == rootPath {
				projectName = filepath.Base(rootPath)
			}

			beats, err := LoadBeats(path)
			if err != nil {
				return nil
			}

			projects = append(projects, model.Project{
				Name:      projectName,
				Path:      path,
				BeatCount: len(beats),
			})
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("discovering projects: %w", err)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].BeatCount > projects[j].BeatCount
	})

	return projects, nil
}

func LoadAllBeats(rootPath string) ([]model.Beat, map[string]string, error) {
	projects, err := DiscoverProjects(rootPath)
	if err != nil {
		return nil, nil, err
	}

	var allBeats []model.Beat
	beatToProject := make(map[string]string)

	for _, proj := range projects {
		beats, err := LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			beatToProject[b.ID] = proj.Name
		}
		allBeats = append(allBeats, beats...)
	}

	sort.Slice(allBeats, func(i, j int) bool {
		return allBeats[i].CreatedAt.After(allBeats[j].CreatedAt)
	})

	return allBeats, beatToProject, nil
}

func SearchBeats(beats []model.Beat, query string) []model.Beat {
	if query == "" {
		return beats
	}
	query = strings.ToLower(query)
	var results []model.Beat
	for _, b := range beats {
		if strings.Contains(strings.ToLower(b.Content), query) ||
			strings.Contains(strings.ToLower(b.Impetus.Label), query) ||
			strings.Contains(strings.ToLower(b.ID), query) {
			results = append(results, b)
		}
	}
	return results
}

func GetDefaultRoot() string {
	if root := os.Getenv("BEATS_ROOT"); root != "" {
		return root
	}
	// Default to current directory - walks up to find .beats
	return "."
}

func FindBeatByID(beats []model.Beat, id string) *model.Beat {
	for i := range beats {
		if beats[i].ID == id {
			return &beats[i]
		}
	}
	return nil
}
