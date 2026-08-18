package usecases

import (
	"errors"
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

func (uc *InitUseCase) Execute(targetDir string) (resultErr error) {
	sddDir := filepath.Join(targetDir, ".sdd")
	if _, err := os.Stat(sddDir); err == nil {
		return fmt.Errorf(".sdd directory already exists in target path")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect target .sdd directory: %w", err)
	}

	subFS, err := fs.Sub(embeds.DefaultSDDResources, "default_sdd")
	if err != nil {
		return fmt.Errorf("failed to access embedded default_sdd: %w", err)
	}
	stageDir, err := os.MkdirTemp(targetDir, ".sdd-init-")
	if err != nil {
		return fmt.Errorf("failed to create .sdd staging directory: %w", err)
	}
	stageOwned := true
	defer func() {
		if stageOwned {
			resultErr = errors.Join(resultErr, os.RemoveAll(stageDir))
		}
	}()

	err = fs.WalkDir(subFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "work-items" || filepath.HasPrefix(path, "work-items/") ||
			path == "tests" || filepath.HasPrefix(path, "tests/") {
			return nil
		}

		targetPath := filepath.Join(stageDir, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return writeInitializedFile(targetPath, data)
	})
	if err != nil {
		return fmt.Errorf("failed to unpack .sdd template: %w", err)
	}

	activeDir := filepath.Join(stageDir, "work-items", "active")
	archiveDir := filepath.Join(stageDir, "work-items", "archive")
	if err := os.MkdirAll(activeDir, 0755); err != nil {
		return fmt.Errorf("failed to create work-items/active: %w", err)
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create work-items/archive: %w", err)
	}

	if err := os.Rename(stageDir, sddDir); err != nil {
		return fmt.Errorf("failed to publish .sdd directory: %w", err)
	}
	stageOwned = false
	return nil
}

func writeInitializedFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		closeErr := file.Close()
		return errors.Join(err, closeErr)
	}
	if err := file.Sync(); err != nil {
		closeErr := file.Close()
		return errors.Join(err, closeErr)
	}
	return file.Close()
}
