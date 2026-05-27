package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"videoforge/internal/ffmpeg"
	"videoforge/internal/jobs"
	"videoforge/internal/ops"
)

// App is the VideoForge core. Every exported method is bound by Wails and
// becomes callable from the React frontend (see frontend/wailsjs/go/main/App).
type App struct {
	ctx  context.Context
	jobs *jobs.Manager
}

func NewApp() *App {
	return &App{}
}

// startup wires the Wails context into the job manager so it can emit events.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.jobs = jobs.NewManager(3, func(event string, data ...interface{}) {
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

func (a *App) SelectOutputPath(defaultName string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save result as",
		DefaultFilename: defaultName,
	})
}

// --- Operations (each starts a job and returns its ID immediately) ---

func (a *App) Convert(p ops.ConvertParams) (string, error) {
	if p.InputPath == "" || p.OutputPath == "" {
		return "", fmt.Errorf("input and output paths are required")
	}
	duration := ffmpeg.Duration(p.InputPath)
	return a.jobs.Start("convert", p.InputPath, p.OutputPath, ops.ConvertArgs(p), duration), nil
}

// --- Reads / control ---

func (a *App) GetMetadata(path string) (*ffmpeg.MediaInfo, error) {
	return ffmpeg.Probe(path)
}

func (a *App) GetJob(id string) jobs.Job {
	return a.jobs.Get(id)
}

func (a *App) CancelJob(id string) {
	a.jobs.Cancel(id)
}

// RevealInFolder opens the OS file manager at the file's containing folder.
func (a *App) RevealInFolder(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+filepath.Dir(path))
}

// SuggestOutputName proposes a default save name for a chosen input + extension.
func (a *App) SuggestOutputName(inputPath, ext string) string {
	base := filepath.Base(inputPath)
	name := base[:len(base)-len(filepath.Ext(base))]
	return name + "_converted" + ext
}
