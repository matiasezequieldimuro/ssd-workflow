package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"sdd-cli/internal/domain"
)

func TestParseValidFixtures(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "..", ".sdd", "tests", "fixtures", "valid")
	files, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to read fixtures dir: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("No fixture files found in %s", fixturesDir)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(fixturesDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", path, err)
			}

			var item domain.WorkItem
			if err := json.Unmarshal(data, &item); err != nil {
				t.Fatalf("Failed to unmarshal fixture %s into domain.WorkItem: %v", file.Name(), err)
			}

			if item.Kind != "work-item" {
				t.Errorf("Expected kind 'work-item', got '%s'", item.Kind)
			}
			if item.ID == "" {
				t.Errorf("Expected non-empty ID")
			}
			if item.Status == "" {
				t.Errorf("Expected non-empty Status")
			}
		})
	}
}
