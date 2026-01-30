// pkg/clients/objectstorage/s3_storage_test.go
package objectstorage

import "testing"

func TestJoinKey(t *testing.T) {
	tests := []struct {
		prefix string
		key    string
		want   string
	}{
		{"", "a/b", "a/b"},
		{"p", "a/b", "p/a/b"},
		{"p/", "a/b", "p/a/b"},
		{"p///", "/a/b", "p/a/b"},
		{"p", "/a/b", "p/a/b"},
		{"p/", "/a/b", "p/a/b"},
	}

	for _, tt := range tests {
		got := JoinKey(tt.prefix, tt.key)
		if got != tt.want {
			t.Fatalf("JoinKey(%q, %q) = %q, want %q", tt.prefix, tt.key, got, tt.want)
		}
	}
}

func TestNormalizeConfig(t *testing.T) {
	// Defaults region when empty
	cfg, err := normalizeConfig(Config{
		Bucket: "b",
		Region: "",
	})
	if err != nil {
		t.Fatalf("normalizeConfig unexpected error: %v", err)
	}
	if cfg.Region != "us-east-1" {
		t.Fatalf("expected default region us-east-1, got %q", cfg.Region)
	}

	// Requires bucket
	_, err = normalizeConfig(Config{})
	if err == nil {
		t.Fatalf("expected error when bucket is missing")
	}

	// Endpoint must include scheme when set
	_, err = normalizeConfig(Config{
		Bucket:   "b",
		Region:   "us-east-1",
		Endpoint: "s3.example.com",
	})
	if err == nil {
		t.Fatalf("expected error when endpoint has no scheme")
	}
}
