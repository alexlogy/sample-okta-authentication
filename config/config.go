package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sample-okta-authentication/models"
)

func loadFromPath(configPath string) (*models.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("failed to read config file", "path", configPath, "error", err)
		return nil, fmt.Errorf("read config file %s: %w", configPath, err)
	}

	var cfg models.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		slog.Error("failed to parse config file", "path", configPath, "error", err)
		return nil, fmt.Errorf("parse config file %s: %w", configPath, err)
	}

	slog.Info("config loaded successfully", "path", configPath)
	return &cfg, nil
}

func Load() (*models.Config, error) {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		slog.Info("loading config from CONFIG_PATH", "path", configPath)
		return loadFromPath(configPath)
	}

	workingDir, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(workingDir, "config.json")
		if _, statErr := os.Stat(configPath); statErr == nil {
			slog.Info("loading config from working directory", "path", configPath)
			return loadFromPath(configPath)
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		slog.Error("failed to resolve executable path", "error", err)
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "config.json")
	slog.Info("loading config from executable directory", "path", configPath)
	return loadFromPath(configPath)
}

func Validate(c *models.Config) error {
	if c == nil {
		err := fmt.Errorf("config is nil")
		slog.Error("config validation failed", "error", err)
		return err
	}
	if err := validateRequiredFile(c.SAMLSPCertFile, "SAML SP cert file"); err != nil {
		slog.Error("config validation failed", "field", "saml_sp_cert_file", "error", err)
		return err
	}
	if err := validateRequiredFile(c.SAMLSPKeyFile, "SAML SP key file"); err != nil {
		slog.Error("config validation failed", "field", "saml_sp_key_file", "error", err)
		return err
	}
	return nil
}

func validateRequiredFile(path string, label string) error {
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
