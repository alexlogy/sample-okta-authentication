package config

import (
	"fmt"
	"log/slog"
	"os"
)

func ValidateRequiredFile(path string, label string) error {
	if path == "" {
		err := fmt.Errorf("%s is not configured", label)
		slog.Error("required file validation failed", "label", label, "error", err)
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := fmt.Errorf("%s does not exist: %s", label, path)
			slog.Error("required file validation failed", "label", label, "path", path, "error", err)
			return err
		}
		wrappedErr := fmt.Errorf("stat %s (%s): %w", label, path, err)
		slog.Error("required file validation failed", "label", label, "path", path, "error", wrappedErr)
		return wrappedErr
	}

	if info.IsDir() {
		err := fmt.Errorf("%s points to a directory, not a file: %s", label, path)
		slog.Error("required file validation failed", "label", label, "path", path, "error", err)
		return err
	}

	return nil
}
