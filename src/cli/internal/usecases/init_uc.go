package usecases

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sdd-cli/embeds"
)

type InitUseCase struct{}

func NewInitUseCase() *InitUseCase {
	return &InitUseCase{}
}

func (uc *InitUseCase) Execute(targetDir string) error {
	sddDir := filepath.Join(targetDir, ".sdd")
	if _, err := os.Stat(sddDir); err == nil {
		return fmt.Errorf(".sdd directory already exists in target path")
	}

	// Copy embedded default_sdd files into targetDir/.sdd
	subFS, err := fs.Sub(embeds.DefaultSDDResources, "default_sdd")
	if err != nil {
		return fmt.Errorf("failed to access embedded default_sdd: %w", err)
	}

	err = fs.WalkDir(subFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip work-items and tests directories from embeds
		if path == "work-items" || filepath.HasPrefix(path, "work-items/") ||
			path == "tests" || filepath.HasPrefix(path, "tests/") {
			return nil
		}

		targetPath := filepath.Join(sddDir, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return os.WriteFile(targetPath, data, 0644)
	})
	if err != nil {
		return fmt.Errorf("failed to unpack .sdd template: %w", err)
	}

	// Ensure work-items active and archive directories exist
	activeDir := filepath.Join(sddDir, "work-items", "active")
	archiveDir := filepath.Join(sddDir, "work-items", "archive")
	if err := os.MkdirAll(activeDir, 0755); err != nil {
		return fmt.Errorf("failed to create work-items/active: %w", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create work-items/archive: %w", err)
	}

	return nil
}
