package env

import (
	"os"
	"strconv"
	"time"
)

func GetString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func GetBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if v == "" {
			return true // treat empty as true when present
		}
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func GetInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func GetDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
