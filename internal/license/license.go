// Package license manages activation against Lemon Squeezy's License API.
//
// LS is the source of truth: this package activates a key once, caches the
// returned {key, instance_id, status, ...} locally, and revalidates on startup.
// If revalidation fails because the network is unreachable, the cached state
// remains valid for OfflineGrace; permanent rejections (key disabled, instance
// deactivated, expired) clear the cache and bounce the user back to the gate.
package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OfflineGrace is how long the app will keep running on cached state when
// Lemon Squeezy is unreachable. Past this, the app forces re-validation.
const OfflineGrace = 14 * 24 * time.Hour

// apiBase is overridable for tests; production always points at LS.
var apiBase = "https://api.lemonsqueezy.com"

// Status mirrors what the UI needs to render the gate, sidebar, and Settings.
type Status struct {
	Activated   bool   `json:"activated"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Status      string `json:"status,omitempty"`      // "active" | "inactive" | "expired" | "disabled"
	KeyShort    string `json:"keyShort,omitempty"`    // safe-to-display prefix from LS
	Expiry      string `json:"expiry,omitempty"`      // YYYY-MM-DD, empty = perpetual
	Perpetual   bool   `json:"perpetual"`
	DaysLeft    int    `json:"daysLeft"`
	Activations int    `json:"activations"`           // instances_count
	Limit       int    `json:"limit"`                 // activation_limit (0 = unlimited)
	Offline     bool   `json:"offline"`               // running on cached state because LS was unreachable
}

// errPermanent is wrapped around any LS rejection that should clear the cache
// (key not valid, expired, disabled, instance gone). Network errors are not
// permanent and trigger the offline grace path instead.
var errPermanent = errors.New("permanent rejection")

// Activate activates a license key with LS and persists the resulting state.
func Activate(key string) (Status, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return Status{}, errors.New("license key is required")
	}
	resp, err := postLicense("/v1/licenses/activate", url.Values{
		"license_key":   {key},
		"instance_name": {instanceName()},
	})
	if err != nil {
		return Status{}, err
	}
	if !resp.Activated {
		return Status{}, errors.New(humanError(resp.Error, "activation failed"))
	}
	s := statusFrom(resp, key)
	st := state{
		Key:           key,
		InstanceID:    resp.Instance.ID,
		Name:          s.Name,
		Email:         s.Email,
		Status:        s.Status,
		KeyShort:      s.KeyShort,
		ExpiresAt:     parseExpiry(resp.LicenseKey.ExpiresAt),
		ActivationLim: resp.LicenseKey.ActivationLimit,
		Instances:     resp.LicenseKey.ActivationUsage,
		LastValidated: time.Now().UTC(),
	}
	if err := saveState(st); err != nil {
		return Status{}, fmt.Errorf("save license state: %w", err)
	}
	return s, nil
}

// LoadStatus returns the current activation state. On a fresh launch with a
// cached key, it revalidates against LS; on network failure it falls back to
// the cached state within OfflineGrace, and otherwise asks the user to re-activate.
func LoadStatus() Status {
	st, ok := loadState()
	if !ok {
		return Status{Activated: false}
	}
	fresh, err := validate(st.Key, st.InstanceID)
	if err == nil {
		merged := mergeFresh(st, fresh)
		_ = saveState(merged)
		return statusFromState(merged, false)
	}
	if errors.Is(err, errPermanent) {
		_ = clearState()
		return Status{Activated: false}
	}
	// Network / transient error.
	if time.Since(st.LastValidated) > OfflineGrace {
		return Status{Activated: false}
	}
	return statusFromState(st, true)
}

// Deactivate releases this device's slot on LS, then clears the local state.
// Best-effort: even if LS is unreachable, the local cache is cleared so the
// user is never stuck.
func Deactivate() error {
	st, ok := loadState()
	if !ok {
		return nil
	}
	_, _ = postLicense("/v1/licenses/deactivate", url.Values{
		"license_key": {st.Key},
		"instance_id": {st.InstanceID},
	})
	return clearState()
}

// --- internal: HTTP + response parsing ---

type lsResponse struct {
	Activated   bool         `json:"activated,omitempty"`
	Valid       bool         `json:"valid,omitempty"`
	Deactivated bool         `json:"deactivated,omitempty"`
	Error       string       `json:"error,omitempty"`
	LicenseKey  lsLicenseKey `json:"license_key"`
	Instance    lsInstance   `json:"instance"`
	Meta        lsMeta       `json:"meta"`
}

type lsLicenseKey struct {
	ID              int    `json:"id"`
	Status          string `json:"status"`
	Key             string `json:"key"`
	KeyShort        string `json:"key_short"`
	ActivationLimit int    `json:"activation_limit"`
	ActivationUsage int    `json:"activation_usage"`
	ExpiresAt       string `json:"expires_at"`
}

type lsInstance struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type lsMeta struct {
	StoreID       int    `json:"store_id"`
	OrderID       int    `json:"order_id"`
	ProductID     int    `json:"product_id"`
	ProductName   string `json:"product_name"`
	VariantID     int    `json:"variant_id"`
	VariantName   string `json:"variant_name"`
	CustomerID    int    `json:"customer_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}

func postLicense(path string, body url.Values) (lsResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodPost, apiBase+path, strings.NewReader(body.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return lsResponse{}, err // network/transient — caller treats as offline
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var r lsResponse
	if jerr := json.Unmarshal(raw, &r); jerr != nil {
		// LS returned non-JSON. Treat as network/transient.
		return lsResponse{}, fmt.Errorf("license api: bad response (%d)", res.StatusCode)
	}
	// LS returns 4xx with {error: "..."} for bad keys; signal as permanent.
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return r, fmt.Errorf("%w: %s", errPermanent, humanError(r.Error, "license rejected"))
	}
	return r, nil
}

func validate(key, instanceID string) (lsResponse, error) {
	resp, err := postLicense("/v1/licenses/validate", url.Values{
		"license_key": {key},
		"instance_id": {instanceID},
	})
	if err != nil {
		return resp, err
	}
	if !resp.Valid {
		// 200 but valid=false → key was disabled/expired/instance removed.
		return resp, fmt.Errorf("%w: %s", errPermanent, humanError(resp.Error, "license no longer valid"))
	}
	return resp, nil
}

// --- internal: helpers ---

func instanceName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return "VideoForge on " + host
}

func parseExpiry(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func statusFrom(r lsResponse, key string) Status {
	expiry := parseExpiry(r.LicenseKey.ExpiresAt)
	return Status{
		Activated:   true,
		Name:        r.Meta.CustomerName,
		Email:       r.Meta.CustomerEmail,
		Status:      r.LicenseKey.Status,
		KeyShort:    pickKeyShort(r.LicenseKey.KeyShort, key),
		Expiry:      formatExpiry(expiry),
		Perpetual:   expiry.IsZero(),
		DaysLeft:    daysLeft(expiry),
		Activations: r.LicenseKey.ActivationUsage,
		Limit:       r.LicenseKey.ActivationLimit,
	}
}

func mergeFresh(st state, fresh lsResponse) state {
	st.Status = fresh.LicenseKey.Status
	st.KeyShort = pickKeyShort(fresh.LicenseKey.KeyShort, st.Key)
	st.ExpiresAt = parseExpiry(fresh.LicenseKey.ExpiresAt)
	st.ActivationLim = fresh.LicenseKey.ActivationLimit
	st.Instances = fresh.LicenseKey.ActivationUsage
	if fresh.Meta.CustomerName != "" {
		st.Name = fresh.Meta.CustomerName
	}
	if fresh.Meta.CustomerEmail != "" {
		st.Email = fresh.Meta.CustomerEmail
	}
	st.LastValidated = time.Now().UTC()
	return st
}

func statusFromState(st state, offline bool) Status {
	return Status{
		Activated:   true,
		Name:        st.Name,
		Email:       st.Email,
		Status:      st.Status,
		KeyShort:    st.KeyShort,
		Expiry:      formatExpiry(st.ExpiresAt),
		Perpetual:   st.ExpiresAt.IsZero(),
		DaysLeft:    daysLeft(st.ExpiresAt),
		Activations: st.Instances,
		Limit:       st.ActivationLim,
		Offline:     offline,
	}
}

func formatExpiry(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func daysLeft(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	d := int(math.Ceil(time.Until(t).Hours() / 24))
	if d < 0 {
		return 0
	}
	return d
}

func pickKeyShort(short, full string) string {
	if short != "" {
		return short
	}
	// Fallback: derive a short form from the full key.
	if len(full) >= 8 {
		return full[:8] + "…"
	}
	return full
}

func humanError(raw, fallback string) string {
	if raw == "" {
		return fallback
	}
	return raw
}
