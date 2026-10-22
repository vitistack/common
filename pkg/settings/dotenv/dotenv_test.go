package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDir creates a test directory and changes to it, returning a cleanup function
func setupTestDir(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir = t.TempDir()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cleanup = func() {
		_ = os.Chdir(origWd)
	}
	return tmpDir, cleanup
}

func TestLoadDotEnv_BasicFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create .env file
	envContent := "TEST_VAR1=value1\nTEST_VAR2=value2\n"
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(envPath) }()

	// Clean environment
	_ = os.Unsetenv("TEST_VAR1")
	_ = os.Unsetenv("TEST_VAR2")
	defer func() { _ = os.Unsetenv("TEST_VAR1") }()
	defer func() { _ = os.Unsetenv("TEST_VAR2") }()

	LoadDotEnv()

	if val := os.Getenv("TEST_VAR1"); val != "value1" {
		t.Errorf("TEST_VAR1 = %q, want %q", val, "value1")
	}
	if val := os.Getenv("TEST_VAR2"); val != "value2" {
		t.Errorf("TEST_VAR2 = %q, want %q", val, "value2")
	}
}

func TestLoadDotEnv_NoOverride(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	envContent := "EXISTING_VAR=from_file\n"
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(envPath) }()

	// Set environment variable before loading
	_ = os.Setenv("EXISTING_VAR", "from_os")
	defer func() { _ = os.Unsetenv("EXISTING_VAR") }()

	LoadDotEnv()

	// Should keep OS value, not override with file value
	if val := os.Getenv("EXISTING_VAR"); val != "from_os" {
		t.Errorf("EXISTING_VAR = %q, want %q (should not override)", val, "from_os")
	}
}

func TestLoadDotEnv_EnvSpecific(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create .env
	baseEnvContent := "BASE_VAR=base_value\n"
	baseEnvPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(baseEnvPath) }()

	// Create .env-test
	testEnvContent := "TEST_ENV_VAR=test_value\n"
	testEnvPath := filepath.Join(tmpDir, ".env-test")
	if err := os.WriteFile(testEnvPath, []byte(testEnvContent), 0644); err != nil {
		t.Fatalf("Failed to create .env-test file: %v", err)
	}
	defer func() { _ = os.Remove(testEnvPath) }()

	// Set ENV
	_ = os.Setenv("ENV", "test")
	defer func() { _ = os.Unsetenv("ENV") }()

	// Clean test variables
	_ = os.Unsetenv("BASE_VAR")
	_ = os.Unsetenv("TEST_ENV_VAR")
	defer func() { _ = os.Unsetenv("BASE_VAR") }()
	defer func() { _ = os.Unsetenv("TEST_ENV_VAR") }()

	LoadDotEnv()

	if val := os.Getenv("BASE_VAR"); val != "base_value" {
		t.Errorf("BASE_VAR = %q, want %q", val, "base_value")
	}
	if val := os.Getenv("TEST_ENV_VAR"); val != "test_value" {
		t.Errorf("TEST_ENV_VAR = %q, want %q", val, "test_value")
	}
}

func TestLoadDotEnv_EnvSpecificOverride(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create .env with OVERRIDE_VAR
	baseEnvContent := "OVERRIDE_VAR=base\n"
	baseEnvPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(baseEnvPath) }()

	// Create .env-prod with same var
	prodEnvContent := "OVERRIDE_VAR=production\n"
	prodEnvPath := filepath.Join(tmpDir, ".env-prod")
	if err := os.WriteFile(prodEnvPath, []byte(prodEnvContent), 0644); err != nil {
		t.Fatalf("Failed to create .env-prod file: %v", err)
	}
	defer func() { _ = os.Remove(prodEnvPath) }()

	// Set ENV
	_ = os.Setenv("ENV", "prod")
	defer func() { _ = os.Unsetenv("ENV") }()

	// Clean test variable
	_ = os.Unsetenv("OVERRIDE_VAR")
	defer func() { _ = os.Unsetenv("OVERRIDE_VAR") }()

	LoadDotEnv()

	// Should use production value
	if val := os.Getenv("OVERRIDE_VAR"); val != "production" {
		t.Errorf("OVERRIDE_VAR = %q, want %q (env-specific should override base)", val, "production")
	}
}

func TestLoadDotEnv_MissingFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Ensure .env doesn't exist
	envPath := filepath.Join(tmpDir, ".env")
	_ = os.Remove(envPath)

	// Should not panic
	LoadDotEnv()
}

func TestLoadDotEnv_MissingEnvSpecific(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create only .env
	baseEnvContent := "BASE_ONLY=value\n"
	baseEnvPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(baseEnvPath) }()

	// Set ENV but don't create .env-{ENV}
	_ = os.Setenv("ENV", "nonexistent")
	defer func() { _ = os.Unsetenv("ENV") }()

	_ = os.Unsetenv("BASE_ONLY")
	defer func() { _ = os.Unsetenv("BASE_ONLY") }()

	// Should not panic, should load .env only
	LoadDotEnv()

	if val := os.Getenv("BASE_ONLY"); val != "value" {
		t.Errorf("BASE_ONLY = %q, want %q", val, "value")
	}
}

func TestFindUpwards_CurrentDir(t *testing.T) {
	// Create a nested directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create file in subDir
	testFile := filepath.Join(subDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() { _ = os.Remove(testFile) }()

	found, ok := findUpwards(subDir, "test.txt")
	if !ok {
		t.Errorf("findUpwards() should find file in current directory")
	}
	if !filepath.IsAbs(found) {
		t.Errorf("findUpwards() should return absolute path, got %q", found)
	}
}

func TestFindUpwards_ParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create file in parent directory
	testFile := filepath.Join(tmpDir, "level1", "parent.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() { _ = os.Remove(testFile) }()

	// Search from deeper directory
	found, ok := findUpwards(subDir, "parent.txt")
	if !ok {
		t.Errorf("findUpwards() should find file in parent directory")
	}
	if found == "" {
		t.Errorf("findUpwards() returned empty path")
	}
}

func TestFindUpwards_RootOfTree(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create file at root
	testFile := filepath.Join(tmpDir, "root.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() { _ = os.Remove(testFile) }()

	// Search from deepest directory
	found, ok := findUpwards(subDir, "root.txt")
	if !ok {
		t.Errorf("findUpwards() should find file at root of search tree")
	}
	if found == "" {
		t.Errorf("findUpwards() returned empty path")
	}
}

func TestFindUpwards_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	_, ok := findUpwards(subDir, "does-not-exist.txt")
	if ok {
		t.Errorf("findUpwards() should return false for nonexistent file")
	}
}

func TestFindFileIfExists_WorkingDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create file in working directory
	testFile := filepath.Join(tmpDir, "wd-test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() { _ = os.Remove(testFile) }()

	found, ok := findFileIfExists("wd-test.txt")
	if !ok {
		t.Errorf("findFileIfExists() should find file in working directory")
	}
	if found == "" {
		t.Errorf("findFileIfExists() returned empty path")
	}
}

func TestFindFileIfExists_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	_, ok := findFileIfExists("totally-nonexistent-file.txt")
	if ok {
		t.Errorf("findFileIfExists() should return false for nonexistent file")
	}
}
