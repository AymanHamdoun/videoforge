package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"videoforge/internal/ffmpeg"
	"videoforge/internal/jobs"
	"videoforge/internal/license"
	"videoforge/internal/ops"
	"videoforge/internal/settings"
)

// Version is the application version, surfaced in the UI via AppVersion().
const Version = "0.1.0"

// PurchaseURL is where the "Buy a license" button sends users. Replace with your
// real store/checkout link (e.g. a Lemon Squeezy / Gumroad product URL).
const PurchaseURL = "https://videoforge.app/buy"

// App is the VideoForge core. Every exported method is bound by Wails and
// becomes callable from the React frontend (see frontend/wailsjs/go/main/App).
type App struct {
	ctx   context.Context
	jobs  *jobs.Manager
	prefs settings.Preferences
}

func NewApp() *App {
	return &App{}
}

// startup wires the Wails context into the job manager so it can emit events.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.prefs = settings.Load()
	a.jobs = jobs.NewManager(a.prefs.MaxConcurrentJobs, func(event string, data ...interface{}) {
		runtime.EventsEmit(ctx, event, data...)
	})
}

// --- File selection (native dialogs; no upload step) ---

var videoFilter = runtime.FileFilter{
	DisplayName: "Video",
	Pattern:     "*.mp4;*.mkv;*.mov;*.avi;*.webm;*.flv;*.wmv;*.ts;*.m4v;*.3gp;*.ogv",
}

func (a *App) SelectInputFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Choose a video",
		Filters: []runtime.FileFilter{videoFilter},
	})
}

func (a *App) SelectInputFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Choose videos to merge (in order)",
		Filters: []runtime.FileFilter{videoFilter},
	})
}

func (a *App) SelectImageFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose a watermark image",
		Filters: []runtime.FileFilter{
			{DisplayName: "Image", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.webp;*.bmp"},
		},
	})
}

func (a *App) SelectOutputPath(defaultName string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Save result as",
		DefaultFilename:  defaultName,
		DefaultDirectory: a.prefs.DefaultOutputDir, // empty = OS default
	})
}

// --- Operations (each starts a job and returns its ID immediately) ---

func (a *App) Convert(p ops.ConvertParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("convert", p.InputPath, p.OutputPath, jobs.Command(ops.ConvertArgs(p), dur)), nil
}

func (a *App) ChangeSpeed(p ops.SpeedParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	if p.Speed <= 0 {
		return "", fmt.Errorf("speed must be greater than 0")
	}
	dur := ffmpeg.Duration(p.InputPath) / p.Speed // output is shorter/longer
	return a.jobs.Start("speed", p.InputPath, p.OutputPath, jobs.Command(ops.SpeedArgs(p), dur)), nil
}

func (a *App) Trim(p ops.TrimParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := parseTimecode(p.Duration)
	return a.jobs.Start("trim", p.InputPath, p.OutputPath, jobs.Command(ops.TrimArgs(p), dur)), nil
}

func (a *App) Compress(p ops.CompressParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("compress", p.InputPath, p.OutputPath, jobs.Command(ops.CompressArgs(p), dur)), nil
}

func (a *App) ExtractAudio(p ops.ExtractParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("extract", p.InputPath, p.OutputPath, jobs.Command(ops.ExtractArgs(p), dur)), nil
}

func (a *App) Rotate(p ops.RotateParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("rotate", p.InputPath, p.OutputPath, jobs.Command(ops.RotateArgs(p), dur)), nil
}

func (a *App) AddWatermark(p ops.WatermarkParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	if p.WatermarkPath == "" {
		return "", fmt.Errorf("a watermark image is required")
	}
	dur := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("watermark", p.InputPath, p.OutputPath, jobs.Command(ops.WatermarkArgs(p), dur)), nil
}

func (a *App) Merge(p ops.MergeParams) (string, error) {
	if len(p.InputPaths) < 2 || p.OutputPath == "" {
		return "", fmt.Errorf("at least two input files and an output path are required")
	}
	// Write the concat demuxer filelist to a temp file.
	listPath := filepath.Join(os.TempDir(), "vf-merge-"+uuid.NewString()+".txt")
	if err := os.WriteFile(listPath, []byte(ops.MergeFileList(p.InputPaths)), 0o600); err != nil {
		return "", fmt.Errorf("write filelist: %w", err)
	}
	var total float64
	for _, in := range p.InputPaths {
		total += ffmpeg.Duration(in)
	}
	fn := func(ctx context.Context, onProgress ffmpeg.ProgressCallback) error {
		defer os.Remove(listPath)
		return ffmpeg.Run(ctx, ops.MergeArgs(listPath, p.OutputPath), total, onProgress)
	}
	return a.jobs.Start("merge", p.InputPaths[0], p.OutputPath, fn), nil
}

func (a *App) MakeGif(p ops.GifParams) (string, error) {
	if err := requireIO(p.InputPath, p.OutputPath); err != nil {
		return "", err
	}
	dur := parseTimecode(p.Duration)
	if dur == 0 {
		dur = ffmpeg.Duration(p.InputPath)
	}
	// Two-pass: generate a palette, then render the GIF with it.
	fn := func(ctx context.Context, onProgress ffmpeg.ProgressCallback) error {
		palette := filepath.Join(os.TempDir(), "vf-palette-"+uuid.NewString()+".png")
		defer os.Remove(palette)
		if err := ffmpeg.Run(ctx, ops.GifPaletteArgs(p, palette), 0, nil); err != nil {
			return fmt.Errorf("palette: %w", err)
		}
		return ffmpeg.Run(ctx, ops.GifArgs(p, palette), dur, onProgress)
	}
	return a.jobs.Start("gif", p.InputPath, p.OutputPath, fn), nil
}

// --- Reads / control ---

func (a *App) GetMetadata(path string) (*ffmpeg.MediaInfo, error) {
	return ffmpeg.Probe(path)
}

func (a *App) GetJob(id string) jobs.Job {
	return a.jobs.Get(id)
}

// AppVersion returns the application version for display in the UI.
func (a *App) AppVersion() string {
	return Version
}

// --- Licensing ---

// LicenseStatus is the activation state surfaced to the UI.
type LicenseStatus struct {
	Activated bool   `json:"activated"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Expiry    string `json:"expiry,omitempty"` // YYYY-MM-DD, empty = perpetual
	Perpetual bool   `json:"perpetual"`
	DaysLeft  int    `json:"daysLeft"` // days until expiry (0 if perpetual)
}

func statusFor(lic *license.License) LicenseStatus {
	s := LicenseStatus{Activated: true, Name: lic.Name, Email: lic.Email}
	if lic.Expiry == nil {
		s.Perpetual = true
		return s
	}
	s.Expiry = lic.Expiry.Format("2006-01-02")
	days := int(math.Ceil(time.Until(*lic.Expiry).Hours() / 24))
	if days < 0 {
		days = 0
	}
	s.DaysLeft = days
	return s
}

// LicenseStatus reports whether a valid license is stored.
func (a *App) LicenseStatus() LicenseStatus {
	key := license.Load()
	if key == "" {
		return LicenseStatus{Activated: false}
	}
	lic, err := license.Verify(key)
	if err != nil {
		return LicenseStatus{Activated: false}
	}
	return statusFor(lic)
}

// Activate validates a license key and, if valid, stores it.
func (a *App) Activate(key string) (LicenseStatus, error) {
	lic, err := license.Verify(key)
	if err != nil {
		return LicenseStatus{Activated: false}, err
	}
	if err := license.Save(key); err != nil {
		return LicenseStatus{Activated: false}, err
	}
	return statusFor(lic), nil
}

// Deactivate removes the stored license.
func (a *App) Deactivate() error {
	return license.Clear()
}

// OpenPurchasePage opens the buy page in the default browser.
func (a *App) OpenPurchasePage() {
	runtime.BrowserOpenURL(a.ctx, PurchaseURL)
}

// --- Preferences ---

// GetPreferences returns the current user preferences.
func (a *App) GetPreferences() settings.Preferences {
	return a.prefs
}

// SavePreferences persists preferences and applies them where possible
// (DefaultOutputDir takes effect immediately; MaxConcurrentJobs on restart).
func (a *App) SavePreferences(p settings.Preferences) (settings.Preferences, error) {
	if err := settings.Save(p); err != nil {
		return a.prefs, err
	}
	a.prefs = settings.Load()
	return a.prefs, nil
}

// SelectFolder opens a native directory picker (for the default output folder).
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose default output folder",
	})
}

func (a *App) CancelJob(id string) {
	a.jobs.Cancel(id)
}

// RevealInFolder opens the OS file manager with the file selected/highlighted.
func (a *App) RevealInFolder(path string) {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", path) // reveal & select in Finder
	case "windows":
		// explorer needs the comma form as a single arg; it exits non-zero even
		// on success, so we don't check the error.
		cmd = exec.Command("explorer", "/select,"+filepath.FromSlash(path))
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(path)) // Linux: open the folder
	}
	_ = cmd.Start()
}

// SuggestOutput proposes a default save name: <base><suffix><ext>.
func (a *App) SuggestOutput(inputPath, suffix, ext string) string {
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return name + suffix + ext
}

// --- helpers ---

func requireIO(in, out string) error {
	if in == "" || out == "" {
		return fmt.Errorf("input and output paths are required")
	}
	return nil
}

// parseTimecode converts "SS", "MM:SS", or "HH:MM:SS" (with optional .ms) to
// seconds. Returns 0 if empty/unparseable (disables progress percentage).
func parseTimecode(tc string) float64 {
	tc = strings.TrimSpace(tc)
	if tc == "" {
		return 0
	}
	if !strings.Contains(tc, ":") {
		v, _ := strconv.ParseFloat(tc, 64)
		return v
	}
	parts := strings.Split(tc, ":")
	var secs float64
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0
		}
		secs = secs*60 + v
	}
	return secs
}
