package infra

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"sdd-cli/internal/domain"
)

func TestContractFixtures(t *testing.T) {
	baseDir := contractBaseDir(t)
	validator := NewSchemaValidator()
	workflowRepo := NewFSWorkflowRepository()

	t.Run("valid fixtures", func(t *testing.T) {
		fixtures := fixtureFiles(t, filepath.Join(baseDir, ".sdd", "tests", "fixtures", "valid"))
		for _, fixturePath := range fixtures {
			t.Run(filepath.Base(fixturePath), func(t *testing.T) {
				item := readFixture(t, fixturePath)
				if err := validator.ValidateValue(baseDir, "work-item.schema.json", item); err != nil {
					t.Fatalf("schema validation failed: %v", err)
				}
				workflow, err := workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
				if err != nil {
					t.Fatalf("GetWorkflow() error = %v", err)
				}
				if err := item.ValidateAgainst(workflow); err != nil {
					t.Fatalf("semantic validation failed: %v", err)
				}
			})
		}
	})

	t.Run("invalid fixtures", func(t *testing.T) {
		fixtures := fixtureFiles(t, filepath.Join(baseDir, ".sdd", "tests", "fixtures", "invalid"))
		for _, fixturePath := range fixtures {
			t.Run(filepath.Base(fixturePath), func(t *testing.T) {
				item := readFixture(t, fixturePath)
				err := validator.ValidateValue(baseDir, "work-item.schema.json", item)
				if err == nil {
					workflow, workflowErr := workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
					if workflowErr != nil {
						err = workflowErr
					} else {
						err = item.ValidateAgainst(workflow)
					}
				}
				if err == nil {
					t.Fatal("invalid fixture passed structural and semantic validation")
				}
			})
		}
	})
}

func TestContainedPathRejectsTraversalAndSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	if _, err := containedPath(root, "..", "outside"); !errors.Is(err, domain.ErrInvalidPath) {
		t.Fatalf("containedPath() traversal error = %v, want %v", err, domain.ErrInvalidPath)
	}

	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	if _, err := containedPath(root, "link", "file.md"); !errors.Is(err, domain.ErrInvalidPath) {
		t.Fatalf("containedPath() symlink error = %v, want %v", err, domain.ErrInvalidPath)
	}
}

func TestRepositoryRejectsInvalidIdentifier(t *testing.T) {
	repository := NewFSWorkItemRepository()
	if _, err := repository.WorkItemExists(t.TempDir(), "../../../escape"); !errors.Is(err, domain.ErrInvalidIdentifier) {
		t.Fatalf("WorkItemExists() error = %v, want %v", err, domain.ErrInvalidIdentifier)
	}
}

func TestRepositoryRejectsSymlinkedWorkItemsRoot(t *testing.T) {
	baseDir := t.TempDir()
	sddDir := filepath.Join(baseDir, ".sdd")
	if err := os.MkdirAll(filepath.Join(sddDir, "work-items"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(sddDir, "work-items", "active")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	if _, err := NewFSWorkItemRepository().WorkItemExists(baseDir, "safe-id"); !errors.Is(err, domain.ErrInvalidPath) {
		t.Fatalf("WorkItemExists() error = %v, want %v", err, domain.ErrInvalidPath)
	}
}

func contractBaseDir(t *testing.T) string {
	t.Helper()
	baseDir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}
	return baseDir
}

func fixtureFiles(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	var files []string
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			files = append(files, filepath.Join(directory, entry.Name()))
		}
	}
	if len(files) == 0 {
		t.Fatalf("no JSON fixtures found in %s", directory)
	}
	return files
}

func readFixture(t *testing.T, path string) *domain.WorkItem {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var item domain.WorkItem
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return &item
}
