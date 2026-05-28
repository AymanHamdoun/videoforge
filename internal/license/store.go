package license

import (
	"os"
	"path/filepath"
)

// activation is persisted to the OS user-config dir so it survives restarts.
func storePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "VideoForge")
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(d, "license.key"), nil
}

// Save stores an activated license key.
func Save(key string) error {
	p, err := storePath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(key), 0o600)
}

// Load returns the stored license key ("" if none).
func Load() string {
	p, err := storePath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// Clear removes the stored license (deactivate).
func Clear() error {
	p, err := storePath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
