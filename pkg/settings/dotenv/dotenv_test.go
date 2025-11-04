package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("loads basic .env file", func(t *testing.T) {
		// Create .env file
		envContent := "TEST_VAR1=value1\nTEST_VAR2=value2\n"
		envPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Remove(envPath)

		// Clean environment
		os.Unsetenv("TEST_VAR1")
		os.Unsetenv("TEST_VAR2")
		defer os.Unsetenv("TEST_VAR1")
		defer os.Unsetenv("TEST_VAR2")

		LoadDotEnv()

		if val := os.Getenv("TEST_VAR1"); val != "value1" {
			t.Errorf("TEST_VAR1 = %q, want %q", val, "value1")
		}
		if val := os.Getenv("TEST_VAR2"); val != "value2" {
			t.Errorf("TEST_VAR2 = %q, want %q", val, "value2")
		}
	})

	t.Run("does not override existing env vars", func(t *testing.T) {
		envContent := "EXISTING_VAR=from_file\n"
		envPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Remove(envPath)

		// Set environment variable before loading
		os.Setenv("EXISTING_VAR", "from_os")
		defer os.Unsetenv("EXISTING_VAR")

		LoadDotEnv()

		// Should keep OS value, not override with file value
		if val := os.Getenv("EXISTING_VAR"); val != "from_os" {
			t.Errorf("EXISTING_VAR = %q, want %q (should not override)", val, "from_os")
		}
	})

	t.Run("loads env-specific file when ENV is set", func(t *testing.T) {
		// Create .env
		baseEnvContent := "BASE_VAR=base_value\n"
		baseEnvPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Remove(baseEnvPath)

		// Create .env-test
		testEnvContent := "TEST_ENV_VAR=test_value\n"
		testEnvPath := filepath.Join(tmpDir, ".env-test")
		if err := os.WriteFile(testEnvPath, []byte(testEnvContent), 0644); err != nil {
			t.Fatalf("Failed to create .env-test file: %v", err)
		}
		defer os.Remove(testEnvPath)

		// Set ENV
		os.Setenv("ENV", "test")
		defer os.Unsetenv("ENV")

		// Clean test variables
		os.Unsetenv("BASE_VAR")
		os.Unsetenv("TEST_ENV_VAR")
		defer os.Unsetenv("BASE_VAR")
		defer os.Unsetenv("TEST_ENV_VAR")

		LoadDotEnv()

		if val := os.Getenv("BASE_VAR"); val != "base_value" {
			t.Errorf("BASE_VAR = %q, want %q", val, "base_value")
		}
		if val := os.Getenv("TEST_ENV_VAR"); val != "test_value" {
			t.Errorf("TEST_ENV_VAR = %q, want %q", val, "test_value")
		}
	})

	t.Run("env-specific file overrides base file values", func(t *testing.T) {
		// Create .env with OVERRIDE_VAR
		baseEnvContent := "OVERRIDE_VAR=base\n"
		baseEnvPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Remove(baseEnvPath)

		// Create .env-prod with same var
		prodEnvContent := "OVERRIDE_VAR=production\n"
		prodEnvPath := filepath.Join(tmpDir, ".env-prod")
		if err := os.WriteFile(prodEnvPath, []byte(prodEnvContent), 0644); err != nil {
			t.Fatalf("Failed to create .env-prod file: %v", err)
		}
		defer os.Remove(prodEnvPath)

		// Set ENV
		os.Setenv("ENV", "prod")
		defer os.Unsetenv("ENV")

		// Clean test variable
		os.Unsetenv("OVERRIDE_VAR")
		defer os.Unsetenv("OVERRIDE_VAR")

		LoadDotEnv()

		// Should use production value
		if val := os.Getenv("OVERRIDE_VAR"); val != "production" {
			t.Errorf("OVERRIDE_VAR = %q, want %q (env-specific should override base)", val, "production")
		}
	})

	t.Run("handles missing .env file gracefully", func(t *testing.T) {
		// Ensure .env doesn't exist
		envPath := filepath.Join(tmpDir, ".env")
		os.Remove(envPath)

		// Should not panic
		LoadDotEnv()
	})

	t.Run("handles missing env-specific file gracefully", func(t *testing.T) {
		// Create only .env
		baseEnvContent := "BASE_ONLY=value\n"
		baseEnvPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(baseEnvPath, []byte(baseEnvContent), 0644); err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}
		defer os.Remove(baseEnvPath)

		// Set ENV but don't create .env-{ENV}
		os.Setenv("ENV", "nonexistent")
		defer os.Unsetenv("ENV")

		os.Unsetenv("BASE_ONLY")
		defer os.Unsetenv("BASE_ONLY")

		// Should not panic, should load .env only
		LoadDotEnv()

		if val := os.Getenv("BASE_ONLY"); val != "value" {
			t.Errorf("BASE_ONLY = %q, want %q", val, "value")
		}
	})
}

func TestFindUpwards(t *testing.T) {
	// Create a nested directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	t.Run("finds file in current directory", func(t *testing.T) {
		// Create file in subDir
		testFile := filepath.Join(subDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		found, ok := findUpwards(subDir, "test.txt")
		if !ok {
			t.Errorf("findUpwards() should find file in current directory")
		}
		if !filepath.IsAbs(found) {
			t.Errorf("findUpwards() should return absolute path, got %q", found)
		}
	})

	t.Run("finds file in parent directory", func(t *testing.T) {
		// Create file in parent directory
		testFile := filepath.Join(tmpDir, "level1", "parent.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		// Search from deeper directory
		found, ok := findUpwards(subDir, "parent.txt")
		if !ok {
			t.Errorf("findUpwards() should find file in parent directory")
		}
		if found == "" {
			t.Errorf("findUpwards() returned empty path")
		}
	})

	t.Run("finds file in root of search tree", func(t *testing.T) {
		// Create file at root
		testFile := filepath.Join(tmpDir, "root.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		// Search from deepest directory
		found, ok := findUpwards(subDir, "root.txt")
		if !ok {
			t.Errorf("findUpwards() should find file at root of search tree")
		}
		if found == "" {
			t.Errorf("findUpwards() returned empty path")
		}
	})

	t.Run("returns false for nonexistent file", func(t *testing.T) {
		_, ok := findUpwards(subDir, "does-not-exist.txt")
		if ok {
			t.Errorf("findUpwards() should return false for nonexistent file")
		}
	})
}

func TestFindFileIfExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("finds file from working directory", func(t *testing.T) {
		// Create file in working directory
		testFile := filepath.Join(tmpDir, "wd-test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile)

		found, ok := findFileIfExists("wd-test.txt")
		if !ok {
			t.Errorf("findFileIfExists() should find file in working directory")
		}
		if found == "" {
			t.Errorf("findFileIfExists() returned empty path")
		}
	})

	t.Run("returns false for nonexistent file", func(t *testing.T) {
		_, ok := findFileIfExists("totally-nonexistent-file.txt")
		if ok {
			t.Errorf("findFileIfExists() should return false for nonexistent file")
		}
	})
}
