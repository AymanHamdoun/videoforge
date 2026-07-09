package license

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// state is the cached LS response we persist between launches. It lets the app
// open instantly while a background revalidation happens, and supports the
// OfflineGrace path when LS is unreachable.
type state struct {
	Key           string    `json:"key"`
	InstanceID    string    `json:"instance_id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Status        string    `json:"status"`
	KeyShort      string    `json:"key_short"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	ActivationLim int       `json:"activation_limit"`
	Instances     int       `json:"instances_count"`
	LastValidated time.Time `json:"last_validated_at"`
}

func storePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "VideoForge")
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(d, "license.json"), nil
}

func saveState(s state) error {
	p, err := storePath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}

func loadState() (state, bool) {
	p, err := storePath()
	if err != nil {
		return state{}, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return state{}, false
	}
	var s state
	if err := json.Unmarshal(b, &s); err != nil {
		return state{}, false
	}
	if s.Key == "" || s.InstanceID == "" {
		return state{}, false
	}
	return s, true
}

func clearState() error {
	p, err := storePath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
