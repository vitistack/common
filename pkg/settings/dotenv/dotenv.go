package dotenv

import (
	"fmt"
	"os"

	"maps"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/vitistack/common/pkg/loggers/vlog"
)

// loadDotEnv loads .env and optional .env-<ENV> without overriding existing OS env vars.
func LoadDotEnv() {
	// Determine environment name from ENV variable (if any)
	env := os.Getenv("ENV")

	// Candidate files in load order (lower to higher precedence)
	candidates := []string{".env"}
	if env != "" {
		candidates = append(candidates, fmt.Sprintf(".env-%s", env))
	}

	// Resolve each file by searching upwards from CWD and executable dir; ignore if missing
	// Merge values so that later files override earlier file values, but never override existing OS env
	merged := map[string]string{}
	loadedFrom := []string{}
	for _, f := range candidates {
		// Find file if it exists
		if p, ok := findFileIfExists(f); ok {
			// Read variables from file
			if kv, err := godotenv.Read(p); err == nil {
				// Merge with precedence: later files override earlier file values
				maps.Copy(merged, kv)
				loadedFrom = append(loadedFrom, p)
			}
		}
	}

	// Apply to process env only for variables that are not already set in OS
	for k, v := range merged {
		if _, exists := os.LookupEnv(k); !exists {
			_ = os.Setenv(k, v)
		}
	}

	// Minimal debug: report which dotenv files were used (paths only, no values)
	if len(loadedFrom) > 0 {
		vlog.Infof("dotenv loaded from: %v\n", loadedFrom)
	}
}

// findFileIfExists searches for the given file name starting from useful roots
// (current working directory and executable directory), walking upwards until
// a match is found. Returns the absolute path if found.
func findFileIfExists(name string) (string, bool) {
	roots := []string{}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}

	// Deduplicate roots while preserving order
	seen := map[string]struct{}{}
	uniqueRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		if _, ok := seen[r]; !ok && r != "" {
			seen[r] = struct{}{}
			uniqueRoots = append(uniqueRoots, r)
		}
	}

	for _, root := range uniqueRoots {
		if p, ok := findUpwards(root, name); ok {
			return p, true
		}
	}
	return "", false
}

// findUpwards looks for name starting at startDir and moving up the directory tree.
func findUpwards(startDir, name string) (string, bool) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			if abs, err2 := filepath.Abs(candidate); err2 == nil {
				return abs, true
			}
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			break
		}
		dir = parent
	}
	return "", false
}
