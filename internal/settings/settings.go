// Package settings persists user preferences as JSON in the OS user-config dir
// (alongside the stored license). Safe for concurrent access.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Preferences are the user-adjustable options.
type Preferences struct {
	// DefaultOutputDir, if set, is where Save dialogs open. Empty = OS default.
	DefaultOutputDir string `json:"defaultOutputDir"`
	// MaxConcurrentJobs caps parallel ffmpeg processes (applied at startup).
	MaxConcurrentJobs int `json:"maxConcurrentJobs"`
}

func Default() Preferences {
	return Preferences{DefaultOutputDir: "", MaxConcurrentJobs: 3}
}

var mu sync.Mutex

func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "VideoForge")
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(d, "settings.json"), nil
}

// Load returns saved preferences, falling back to defaults for anything missing.
func Load() Preferences {
	mu.Lock()
	defer mu.Unlock()
	p := Default()
	fp, err := filePath()
	if err != nil {
		return p
	}
	if b, err := os.ReadFile(fp); err == nil {
		_ = json.Unmarshal(b, &p)
	}
	return normalize(p)
}

// Save writes preferences to disk.
func Save(p Preferences) error {
	mu.Lock()
	defer mu.Unlock()
	p = normalize(p)
	fp, err := filePath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fp, b, 0o600)
}

func normalize(p Preferences) Preferences {
	if p.MaxConcurrentJobs < 1 {
		p.MaxConcurrentJobs = 1
	}
	if p.MaxConcurrentJobs > 8 {
		p.MaxConcurrentJobs = 8
	}
	return p
}
