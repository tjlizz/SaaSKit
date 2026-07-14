package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadDotEnv() (string, error) {
	candidates := []string{".env", filepath.Join("..", ".env")}
	if executable, err := os.Executable(); err == nil {
		dir := filepath.Dir(executable)
		candidates = append(candidates, filepath.Join(dir, ".env"), filepath.Join(dir, "..", ".env"))
	}
	return loadFirstDotEnv(candidates...)
}

func loadFirstDotEnv(paths ...string) (string, error) {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		if _, exists := seen[absolute]; exists {
			continue
		}
		seen[absolute] = struct{}{}

		if _, err := os.Stat(absolute); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		// Load intentionally preserves variables already supplied by the
		// process environment, Docker, or the operating system.
		if err := godotenv.Load(absolute); err != nil {
			return "", fmt.Errorf("load %s: %w", absolute, err)
		}
		return absolute, nil
	}
	return "", nil
}
