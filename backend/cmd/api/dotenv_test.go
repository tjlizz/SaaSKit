package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFirstDotEnvLoadsFileWithoutOverridingEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("SAASKIT_DOTENV_FILE=loaded\nSAASKIT_DOTENV_EXISTING=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SAASKIT_DOTENV_EXISTING", "from-process")
	_ = os.Unsetenv("SAASKIT_DOTENV_FILE")
	t.Cleanup(func() { _ = os.Unsetenv("SAASKIT_DOTENV_FILE") })

	loadedPath, err := loadFirstDotEnv(filepath.Join(dir, "missing.env"), path)
	if err != nil {
		t.Fatal(err)
	}
	if loadedPath != path {
		t.Fatalf("want loaded path %q, got %q", path, loadedPath)
	}
	if got := os.Getenv("SAASKIT_DOTENV_FILE"); got != "loaded" {
		t.Fatalf("want value from file, got %q", got)
	}
	if got := os.Getenv("SAASKIT_DOTENV_EXISTING"); got != "from-process" {
		t.Fatalf("process environment was overwritten: %q", got)
	}
}

func TestLoadFirstDotEnvAllowsMissingFile(t *testing.T) {
	loadedPath, err := loadFirstDotEnv(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil || loadedPath != "" {
		t.Fatalf("missing .env should be optional, path=%q err=%v", loadedPath, err)
	}
}
