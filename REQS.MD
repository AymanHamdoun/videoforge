# CLAUDE.md — VideoForge: Native Desktop Video Toolkit

## Project Overview

Build **VideoForge**, a fully offline **desktop application** for video processing. A Go core wraps a **bundled** ffmpeg/ffprobe via `os/exec`. A React frontend (Vite + TailwindCSS v4) provides a drag-and-drop UI. The whole thing ships as a single installable app — `.dmg` on macOS, `.exe`/NSIS installer on Windows — with **zero internet and zero external dependencies required** (ffmpeg travels inside the bundle).

### Distribution Goal

- One double-click installer per platform. User installs it like any native app — no Docker, no Node, no ffmpeg install, no terminal.
- Must **feel fast and native**, not like a browser tab. This rules out Electron's bundled Chromium.

### Why Wails (not Electron/Tauri)

- **Wails v2** uses the OS-native webview (WebKit via WKWebView on macOS, **WebView2** on Windows) — no bundled Chromium. Binaries are ~10–25 MB and start instantly.
- The core is already Go, so Wails reuses it directly: Go methods are **bound** and callable from the React frontend, no HTTP server, no ports, no CORS.
- Tauri would mean a Rust shell + a Go sidecar binary + IPC plumbing — unnecessary overhead for an all-Go backend.

> **Architecture decisions for this build:** desktop-only (no Docker/Nginx), and **native Wails bindings + events** instead of an HTTP API + SSE. The ffmpeg command reference below is unchanged and remains the source of truth for processing logic.

## Monorepo Structure

```
videoforge/
├── wails.json                  # Wails project config (app name, build hooks)
├── go.mod                      # Go 1.23+, module: github.com/videoforge/app
├── go.sum
├── main.go                     # Wails entry: embed frontend, bind App, set options
├── app.go                      # App struct: startup ctx + all bound methods (the "API")
├── internal/
│   ├── ffmpeg/                 # ffmpeg command builder + executor
│   │   ├── runner.go           # os/exec wrapper; reports progress via callback
│   │   ├── probe.go            # ffprobe metadata extraction
│   │   ├── args.go             # Type-safe argument builder
│   │   └── binaries.go         # Resolves bundled ffmpeg/ffprobe paths per-OS
│   ├── jobs/                   # In-memory job queue + status tracking
│   │   └── manager.go
│   └── ops/                    # One file per operation (pure arg-building + run)
│       ├── convert.go
│       ├── speed.go
│       ├── trim.go
│       ├── compress.go
│       ├── extract.go
│       ├── merge.go
│       ├── gif.go
│       ├── watermark.go
│       ├── rotate.go
│       └── metadata.go
├── build/                      # Wails build assets (icons, installer config)
│   ├── bin/                    # BUNDLED ffmpeg + ffprobe binaries, per platform
│   │   ├── darwin/             #   ffmpeg, ffprobe (universal or per-arch)
│   │   └── windows/            #   ffmpeg.exe, ffprobe.exe
│   ├── appicon.png
│   ├── darwin/                 # Info.plist, entitlements, dmg layout
│   └── windows/                # NSIS installer config, icon, app manifest
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── index.html
│   ├── wailsjs/                # AUTO-GENERATED Go bindings + runtime (do not edit)
│   │   ├── go/main/App.d.ts    #   typed wrappers for every bound method
│   │   └── runtime/
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── components/
│       │   ├── Layout.tsx
│       │   ├── FileDropZone.tsx     # Native drag-and-drop (Wails OnFileDrop) → file paths
│       │   ├── JobProgress.tsx      # Progress bar driven by Wails events
│       │   ├── FormatConverter.tsx
│       │   ├── SpeedChanger.tsx
│       │   ├── VideoTrimmer.tsx
│       │   ├── Compressor.tsx
│       │   ├── AudioExtractor.tsx
│       │   ├── VideoMerger.tsx
│       │   ├── GifMaker.tsx
│       │   ├── WatermarkAdder.tsx
│       │   ├── VideoRotator.tsx
│       │   └── MetadataViewer.tsx
│       ├── hooks/
│       │   ├── useJobEvents.ts      # Subscribes to Wails progress events
│       │   └── useFilePicker.ts     # Wraps native open/save dialogs
│       └── types/
│           └── index.ts
└── README.md
```

## Tech Stack & Versions

| Layer | Tech | Version / Notes |
|-------|------|-----------------|
| Desktop shell | **Wails** | **v2** (stable). Native webview, no Chromium. `wails` CLI for dev/build. |
| Core | Go | 1.23+ (no HTTP router needed — methods are bound, not routed) |
| Webview | WKWebView / WebView2 | macOS ships WebKit; Windows requires **WebView2 runtime** (see installer notes) |
| ffmpeg | ffmpeg + ffprobe | **Bundled** in `build/bin/<os>/`. Static builds (e.g. from BtbN / evermeet.cx for macOS, gyan.dev for Windows) |
| Frontend | React 19 + TypeScript | Vite 6 as bundler (Wails wraps it) |
| Frontend CSS | TailwindCSS | v4 (CSS-first via `@import "tailwindcss"`) or v3 if v4 causes issues |
| Packaging | Wails build + NSIS / create-dmg | `.exe` installer on Windows, `.dmg` on macOS |

## Desktop Packaging (Wails)

### wails.json

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "VideoForge",
  "outputfilename": "VideoForge",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": { "name": "VideoForge" }
}
```

### main.go — App entry

```go
package main

import (
    "embed"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "VideoForge",
        Width:  1100,
        Height: 720,
        AssetServer: &assetserver.Options{Assets: assets},
        OnStartup:   app.startup,
        // Enable native file-path drag-and-drop into the window
        DragAndDrop: &options.DragAndDrop{EnableFileDrop: true},
        Bind: []interface{}{
            app, // every exported method on *App becomes callable from JS
        },
    })
    if err != nil {
        panic(err)
    }
}
```

### Bundling ffmpeg (`internal/ffmpeg/binaries.go`)

ffmpeg/ffprobe are **not** assumed to be on the user's machine. Ship static binaries inside the app bundle and resolve their path at runtime.

```go
// Resolve the bundled ffmpeg/ffprobe path.
// On macOS the binaries live in <App>.app/Contents/Resources/bin/.
// On Windows they sit next to the .exe in a bin/ subfolder.
// During `wails dev` they resolve to ./build/bin/<os>/.
func FFmpegPath() string  { return resolveBin("ffmpeg") }
func FFprobePath() string { return resolveBin("ffprobe") }
```

- Add the platform binary folder to the bundle via Wails build resources (`build/darwin` Info.plist + `Resources`, `build/windows` installer copies `bin/`).
- Mark them executable on macOS (`chmod +x` happens in a build hook); they are signed/notarized as part of the app on macOS.
- `runner.go` and `probe.go` call `FFmpegPath()` / `FFprobePath()` instead of the bare strings `"ffmpeg"` / `"ffprobe"`.

### Build commands

```bash
# Dev (hot reload, native window)
wails dev

# macOS — universal .app, then package a .dmg
wails build -platform darwin/universal -clean
# .app lands in build/bin/VideoForge.app
create-dmg "VideoForge.dmg" "build/bin/VideoForge.app"   # or appdmg / hdiutil

# Windows — installer .exe via built-in NSIS support
wails build -platform windows/amd64 -nsis -clean
# installer lands in build/bin/VideoForge-amd64-installer.exe
```

### Distribution notes

- **macOS:** unsigned apps trigger Gatekeeper. For real distribution, sign with a Developer ID cert and notarize (`xcrun notarytool`). Document this in README; dev/local builds can be opened via right-click → Open.
- **Windows:** the **WebView2 runtime** is required. It's preinstalled on Windows 11 and recent Win10; for older machines, bundle the Evergreen Bootstrapper in the NSIS installer (Wails supports `webview2: download/embed` strategies).
- **No volumes / no upload dir:** the app reads input files in place from paths the user picks (or drops), writes to OS temp, then prompts a native **Save** dialog for the result. There is no server-side `uploads/` or `outputs/` directory.

## Core Architecture (Bound Methods, not HTTP)

There is **no HTTP server and no router**. The `App` struct's exported methods are the API; Wails generates typed TS wrappers in `frontend/wailsjs/go/main/App`.

### app.go — bound methods

```go
type App struct {
    ctx  context.Context
    jobs *jobs.Manager
}

func NewApp() *App { return &App{jobs: jobs.NewManager()} }

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// --- File selection (native dialogs, no upload step) ---

func (a *App) SelectInputFile() (string, error) {
    return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
        Title: "Choose a video",
        Filters: []runtime.FileFilter{
            {DisplayName: "Video", Pattern: "*.mp4;*.mkv;*.mov;*.avi;*.webm;*.flv;*.wmv;*.ts;*.m4v;*.3gp;*.ogv"},
        },
    })
}

func (a *App) SelectOutputPath(defaultName string) (string, error) {
    return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{DefaultFilename: defaultName})
}

// --- Operations: each starts a job goroutine, returns a job ID immediately ---

func (a *App) Convert(p ops.ConvertParams) (string, error)        { return a.start("convert", p) }
func (a *App) ChangeSpeed(p ops.SpeedParams) (string, error)      { return a.start("speed", p) }
func (a *App) Trim(p ops.TrimParams) (string, error)              { return a.start("trim", p) }
func (a *App) Compress(p ops.CompressParams) (string, error)      { return a.start("compress", p) }
func (a *App) ExtractAudio(p ops.ExtractParams) (string, error)   { return a.start("extract", p) }
func (a *App) MergeVideos(p ops.MergeParams) (string, error)      { return a.start("merge", p) }
func (a *App) MakeGif(p ops.GifParams) (string, error)            { return a.start("gif", p) }
func (a *App) AddWatermark(p ops.WatermarkParams) (string, error) { return a.start("watermark", p) }
func (a *App) Rotate(p ops.RotateParams) (string, error)          { return a.start("rotate", p) }

// --- Synchronous reads ---

func (a *App) GetMetadata(path string) (*ffmpeg.MediaInfo, error) { return ffmpeg.Probe(path) }
func (a *App) GetJob(id string) jobs.Job                          { return a.jobs.Get(id) }
func (a *App) CancelJob(id string) error                          { return a.jobs.Cancel(id) }
```

Frontend call site (generated bindings):

```ts
import { SelectInputFile, Convert } from "../wailsjs/go/main/App";

const inputPath = await SelectInputFile();        // native picker → absolute path
const jobId     = await Convert({ inputPath, outputPath, crf: 23, /* ... */ });
```

### Progress via Wails Events (replaces SSE)

The `ProgressCallback` in `runner.go` emits a Wails event instead of writing to an SSE stream. The frontend subscribes with `EventsOn` and unsubscribes on unmount.

```go
// inside the job goroutine
onProgress := func(percent float64) {
    a.jobs.SetProgress(id, percent)
    runtime.EventsEmit(a.ctx, "job:progress", jobs.Progress{ID: id, Percent: percent})
}
// on completion / failure:
runtime.EventsEmit(a.ctx, "job:done", jobs.Result{ID: id, Status: "completed", Output: outPath})
```

```ts
// useJobEvents.ts
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";

useEffect(() => {
  const off = EventsOn("job:progress", (p: { id: string; percent: number }) => {
    if (p.id === jobId) setPercent(p.percent);
  });
  return () => { EventsOff("job:progress"); off?.(); };
}, [jobId]);
```

## ffmpeg Command Builder (`internal/ffmpeg/`)

Use `os/exec` directly — do NOT use `ffmpeg-go` or `goffmpeg` libraries. Build commands as string slices for safety. Always invoke the **bundled** binary via `FFmpegPath()` / `FFprobePath()`.

### runner.go — Execute with Progress Parsing

```go
package ffmpeg

import (
    "bufio"
    "context"
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
)

type ProgressCallback func(percent float64)

func Run(ctx context.Context, args []string, totalDuration float64, onProgress ProgressCallback) error {
    // IMPORTANT: Always include these flags for progress parsing:
    // -y (overwrite), -progress pipe:1 -nostats
    fullArgs := append([]string{"-y", "-progress", "pipe:1", "-nostats"}, args...)
    cmd := exec.CommandContext(ctx, FFmpegPath(), fullArgs...) // bundled binary, not bare "ffmpeg"

    stdout, _ := cmd.StdoutPipe()
    cmd.Stderr = nil // suppress stderr noise

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("ffmpeg start: %w", err)
    }

    // Parse progress output: "out_time_us=12345678"
    scanner := bufio.NewScanner(stdout)
    timeRegex := regexp.MustCompile(`out_time_us=(\d+)`)
    for scanner.Scan() {
        line := scanner.Text()
        if matches := timeRegex.FindStringSubmatch(line); len(matches) == 2 {
            us, _ := strconv.ParseFloat(matches[1], 64)
            seconds := us / 1_000_000
            if totalDuration > 0 && onProgress != nil {
                percent := (seconds / totalDuration) * 100
                if percent > 100 { percent = 100 }
                onProgress(percent)
            }
        }
    }

    return cmd.Wait()
}
```

### probe.go — Get Metadata via ffprobe

```go
func Probe(filepath string) (*MediaInfo, error) {
    cmd := exec.Command(FFprobePath(), // bundled binary
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        "-show_streams",
        filepath,
    )
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("ffprobe: %w", err)
    }
    var info MediaInfo
    json.Unmarshal(output, &info)
    return &info, nil
}
```

### Feature-Specific ffmpeg Commands

All commands below are verified against current ffmpeg docs. Build these as `[]string` slices passed to `Run()`.

#### 1. Format Conversion

Supported: mp4, mkv, avi, mov, flv, webm, wmv, ts, m4v, 3gp, ogv

```go
// Basic conversion — ffmpeg auto-selects codecs based on output extension
args := []string{"-i", inputPath, outputPath}

// Explicit codec control for quality
args := []string{
    "-i", inputPath,
    "-c:v", "libx264",    // H.264 video codec
    "-preset", "medium",   // Encoding speed/quality tradeoff: ultrafast|superfast|veryfast|faster|fast|medium|slow|slower|veryslow
    "-crf", "23",          // Quality: 0 (lossless) to 51 (worst), 18-28 is sane range
    "-c:a", "aac",         // AAC audio
    "-b:a", "192k",        // Audio bitrate
    "-pix_fmt", "yuv420p", // CRITICAL: ensures compatibility with most players
    outputPath,
}

// WebM output (VP9 + Opus)
args := []string{
    "-i", inputPath,
    "-c:v", "libvpx-vp9",
    "-crf", "30",
    "-b:v", "0",           // Required for CRF mode with VP9
    "-c:a", "libopus",
    outputPath,
}
```

#### 2. Speed Change (setpts + atempo)

```go
// Speed up 2x — video AND audio in sync
// Video: setpts=0.5*PTS (inverse of speed)
// Audio: atempo=2.0 (direct speed multiplier)
// IMPORTANT: atempo range is [0.5, 100.0], but quality degrades above 2.0
// For >2x audio speed, chain atempo filters: atempo=2.0,atempo=2.0 = 4x
speed := 2.0
videoPTS := fmt.Sprintf("setpts=%f*PTS", 1.0/speed)

// Build atempo chain for audio (each filter max 2.0x)
atempoFilters := buildAtempoChain(speed) // e.g., "atempo=2.0" or "atempo=2.0,atempo=2.0"

args := []string{
    "-i", inputPath,
    "-filter_complex",
    fmt.Sprintf("[0:v]%s[v];[0:a]%s[a]", videoPTS, atempoFilters),
    "-map", "[v]", "-map", "[a]",
    "-c:v", "libx264", "-preset", "fast", "-crf", "22",
    "-c:a", "aac", "-b:a", "192k",
    "-pix_fmt", "yuv420p",
    outputPath,
}

// Helper: chain atempo filters for speeds > 2.0x
func buildAtempoChain(speed float64) string {
    var filters []string
    remaining := speed
    for remaining > 2.0 {
        filters = append(filters, "atempo=2.0")
        remaining /= 2.0
    }
    if remaining < 0.5 {
        // For very slow speeds, chain 0.5x filters
        for remaining < 0.5 {
            filters = append(filters, "atempo=0.5")
            remaining /= 0.5
        }
    }
    filters = append(filters, fmt.Sprintf("atempo=%f", remaining))
    return strings.Join(filters, ",")
}
```

#### 3. Trim / Cut

```go
// Trim with stream copy (fast, no re-encoding)
args := []string{
    "-ss", startTime,    // e.g., "00:01:30" or "90" (seconds)
    "-i", inputPath,
    "-t", duration,      // Duration, e.g., "00:00:30"
    "-c", "copy",        // Copy all streams without re-encoding
    outputPath,
}

// NOTE: Place -ss BEFORE -i for fast seeking (input seeking)
// Place -ss AFTER -i for frame-accurate seeking (slower, re-encodes)

// Trim with re-encoding for frame-accurate cuts
args := []string{
    "-i", inputPath,
    "-ss", startTime,
    "-to", endTime,      // End time (alternative to -t duration)
    "-c:v", "libx264", "-crf", "22",
    "-c:a", "aac",
    "-pix_fmt", "yuv420p",
    outputPath,
}
```

#### 4. Compress / Reduce File Size

```go
// CRF-based compression (visually lossless at 18, good at 23, smaller at 28)
args := []string{
    "-i", inputPath,
    "-c:v", "libx264",
    "-preset", preset,     // "slow" = better compression, "fast" = quicker
    "-crf", crf,           // User-selected: 18-51
    "-c:a", "aac",
    "-b:a", "128k",        // Lower audio bitrate for compression
    "-pix_fmt", "yuv420p",
    outputPath,
}

// Target file size compression (two-pass)
// Pass 1:
args1 := []string{
    "-i", inputPath,
    "-c:v", "libx264", "-preset", "slow",
    "-b:v", targetBitrate, // Calculate: (targetSizeMB * 8192) / durationSeconds - audioBitrate
    "-pass", "1",
    "-an", "-f", "null", "/dev/null",
}
// Pass 2:
args2 := []string{
    "-i", inputPath,
    "-c:v", "libx264", "-preset", "slow",
    "-b:v", targetBitrate,
    "-pass", "2",
    "-c:a", "aac", "-b:a", "128k",
    outputPath,
}
```

> **Note (desktop):** two-pass uses a log file (`ffmpeg2pass-*.log`) and `/dev/null` (`NUL` on Windows). Run passes in a per-job temp dir and clean it up after. Use `os.DevNull` for the cross-platform null sink.

#### 5. Extract Audio

```go
// Extract as MP3
args := []string{
    "-i", inputPath,
    "-vn",                 // Disable video
    "-c:a", "libmp3lame",
    "-q:a", "2",           // VBR quality: 0 (best) to 9 (worst), 2 is ~190kbps
    outputPath,            // .mp3
}

// Extract as WAV (lossless)
args := []string{"-i", inputPath, "-vn", outputPath} // .wav

// Extract as FLAC
args := []string{"-i", inputPath, "-vn", "-c:a", "flac", outputPath}

// Extract as AAC
args := []string{"-i", inputPath, "-vn", "-c:a", "aac", "-b:a", "256k", outputPath}
```

#### 6. Merge / Concatenate Videos

```go
// For same-codec files: concat demuxer (fast, no re-encode)
// First write a filelist.txt:
//   file '/path/to/video1.mp4'
//   file '/path/to/video2.mp4'
args := []string{
    "-f", "concat",
    "-safe", "0",
    "-i", filelistPath,
    "-c", "copy",
    outputPath,
}

// For mixed-codec files: concat filter (re-encodes)
// Must normalize resolution, codec, framerate first
args := []string{
    "-i", input1, "-i", input2,
    "-filter_complex",
    "[0:v]scale=1920:1080,setsar=1[v0];[1:v]scale=1920:1080,setsar=1[v1];[v0][0:a][v1][1:a]concat=n=2:v=1:a=1[v][a]",
    "-map", "[v]", "-map", "[a]",
    "-c:v", "libx264", "-crf", "22",
    "-c:a", "aac",
    "-pix_fmt", "yuv420p",
    outputPath,
}
```

#### 7. Create GIF

```go
// Two-pass GIF for quality (generate palette first)
// Pass 1: Generate palette
args1 := []string{
    "-i", inputPath,
    "-ss", startTime,
    "-t", duration,
    "-vf", fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos,palettegen", fps, width),
    "-y", palettePath,
}
// Pass 2: Use palette for high-quality GIF
args2 := []string{
    "-i", inputPath,
    "-i", palettePath,
    "-ss", startTime,
    "-t", duration,
    "-filter_complex",
    fmt.Sprintf("[0:v]fps=%d,scale=%d:-1:flags=lanczos[v];[v][1:v]paletteuse", fps, width),
    outputPath,
}
```

#### 8. Add Watermark (image overlay)

```go
// Overlay image watermark
// Positions: top-left "10:10", top-right "W-w-10:10", bottom-right "W-w-10:H-h-10"
args := []string{
    "-i", inputPath,
    "-i", watermarkPath,
    "-filter_complex",
    fmt.Sprintf("overlay=%s", position), // e.g., "W-w-10:H-h-10" for bottom-right
    "-c:v", "libx264", "-crf", "22",
    "-c:a", "copy",
    "-pix_fmt", "yuv420p",
    outputPath,
}

// Text watermark
args := []string{
    "-i", inputPath,
    "-vf", fmt.Sprintf(
        "drawtext=text='%s':fontsize=%d:fontcolor=white@0.5:x=(w-text_w)/2:y=h-th-20",
        text, fontSize,
    ),
    "-c:v", "libx264", "-crf", "22",
    "-c:a", "copy",
    outputPath,
}
```

> **Note (desktop):** `drawtext` needs a font file. Bundle a font (e.g. DejaVuSans.ttf) in `build/bin/` and pass `fontfile=<bundled path>` rather than relying on system fonts, which differ across machines.

#### 9. Rotate / Flip

```go
// Rotate 90° clockwise
args := []string{"-i", inputPath, "-vf", "transpose=1", "-c:a", "copy", outputPath}
// Rotate 90° counter-clockwise
args := []string{"-i", inputPath, "-vf", "transpose=2", "-c:a", "copy", outputPath}
// Rotate 180°
args := []string{"-i", inputPath, "-vf", "transpose=1,transpose=1", "-c:a", "copy", outputPath}
// Horizontal flip
args := []string{"-i", inputPath, "-vf", "hflip", "-c:a", "copy", outputPath}
// Vertical flip
args := []string{"-i", inputPath, "-vf", "vflip", "-c:a", "copy", outputPath}
```

#### 10. Resolution / Scale

```go
// Scale to specific resolution (maintain aspect ratio with -2)
args := []string{
    "-i", inputPath,
    "-vf", "scale=1280:-2",  // Width 1280, height auto (divisible by 2)
    "-c:v", "libx264", "-crf", "22",
    "-c:a", "copy",
    "-pix_fmt", "yuv420p",
    outputPath,
}
```

### Job Queue System (`internal/jobs/`)

Use an in-memory job manager with `sync.RWMutex`. Each bound operation creates a job with a UUID, runs ffmpeg in a goroutine, and updates progress via the `ProgressCallback` (which emits Wails events). Store a `context.CancelFunc` per job so `CancelJob` can abort a running ffmpeg process.

```go
type Job struct {
    ID         string    `json:"id"`
    Status     string    `json:"status"`     // queued | processing | completed | failed | canceled
    Progress   float64   `json:"progress"`   // 0-100
    Type       string    `json:"type"`       // convert | speed | trim | etc.
    InputFile  string    `json:"input_file"`
    OutputFile string    `json:"output_file,omitempty"`
    Error      string    `json:"error,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}
```

## Frontend Architecture

### Key Principles

- **TailwindCSS v4**: Use `@import "tailwindcss"` in CSS. If v4 has issues, fall back to v3 with `tailwind.config.ts`.
- **Dark theme by default** with a toggle. Use Tailwind's `dark:` variant.
- **Component per feature**: Each tool (convert, speed, trim, etc.) is its own component with its own form.
- **Files are paths, not uploads**: Use the native open dialog (`SelectInputFile`) or Wails `OnFileDrop` for drag-and-drop. The frontend works with absolute file paths — there is no multipart upload and no `FormData`.
- **Progress via Wails events**: subscribe with `EventsOn("job:progress", ...)`; unsubscribe in `useEffect` cleanup with `EventsOff`.
- **Output via Save dialog**: After processing, the result path is returned and revealed in the OS file manager (`runtime.BrowserOpenURL`/reveal), or chosen up front via `SelectOutputPath`.
- **localStorage IS allowed** here (real packaged app, not a sandboxed artifact) — use it for theme + last-used settings if desired.

### UI Layout

```
┌──────────────────────────────────────────────┐
│  🎬 VideoForge          [Dark/Light Toggle]  │
├──────────────────────────────────────────────┤
│  ┌──────────┐  ┌─────────────────────────┐   │
│  │ Sidebar  │  │  Main Content Area      │   │
│  │          │  │                         │   │
│  │ Convert  │  │  [Drag & Drop Zone]     │   │
│  │ Speed    │  │                         │   │
│  │ Trim     │  │  [Tool-specific form]   │   │
│  │ Compress │  │                         │   │
│  │ Audio    │  │  [Progress bar]         │   │
│  │ Merge    │  │                         │   │
│  │ GIF      │  │  [Reveal / Open output] │   │
│  │ Watermark│  │                         │   │
│  │ Rotate   │  │                         │   │
│  │ Metadata │  │                         │   │
│  └──────────┘  └─────────────────────────┘   │
└──────────────────────────────────────────────┘
```

### Calling the Go core (generated bindings)

```ts
import { SelectInputFile, SelectOutputPath, Convert, GetJob } from "../wailsjs/go/main/App";
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";

async function runConvert() {
  const inputPath  = await SelectInputFile();
  const outputPath = await SelectOutputPath("output.mp4");
  const jobId      = await Convert({ inputPath, outputPath, crf: 23, preset: "medium" });

  const off = EventsOn("job:progress", (p) => {
    if (p.id === jobId) setPercent(p.percent);
  });
  EventsOn("job:done", (r) => {
    if (r.id === jobId) { off?.(); EventsOff("job:progress"); showDone(r.output); }
  });
}
```

## Critical Implementation Rules

### Go Core

1. **No HTTP router, no server** — operations are exported methods on `*App`, bound via Wails. No `gorilla/mux`, `chi`, `gin`, or `net/http` routing.
2. **`os/exec` directly** — do NOT import `ffmpeg-go`, `goffmpeg`, or any wrapper. Build `[]string` arg slices.
3. **Always call the bundled binary** via `FFmpegPath()` / `FFprobePath()` — never the bare `"ffmpeg"`/`"ffprobe"` strings (the user's machine has no system ffmpeg).
4. **Always `-y` flag** — overwrite output files without prompting.
5. **Always `-pix_fmt yuv420p`** — on any re-encoded video output for player compatibility.
6. **Always `-progress pipe:1 -nostats`** — for parseable progress output.
7. **Place `-ss` before `-i`** for fast seeking in trim operations.
8. **Parse `out_time_us`** from ffmpeg progress (microseconds) → percent → `runtime.EventsEmit`.
9. **Use `context.Context` per job** — pass to `exec.CommandContext`; keep the `CancelFunc` for `CancelJob`.
10. **Limit concurrent jobs** — semaphore (buffered channel) to cap parallel ffmpeg processes.
11. **Per-job temp dir** — write intermediates (palettes, two-pass logs) under `os.TempDir()` and delete on completion. Use `os.DevNull`, not a hardcoded `/dev/null`, for cross-platform.
12. **Cross-platform paths** — never hardcode `/`; use `path/filepath`.

### ffmpeg Commands

1. **atempo range**: officially [0.5, 100.0] but quality degrades above 2.0. Chain multiple `atempo=2.0` filters for high-speed audio.
2. **CRF range**: 0-51 for H.264. 18 = visually lossless, 23 = default, 28+ = visible compression.
3. **Presets**: `ultrafast` → `veryslow`. Slower = better compression ratio at same CRF.
4. **GIF**: Always use two-pass with palettegen/paletteuse for quality.
5. **Concat**: Use concat demuxer (`-f concat -c copy`) for same-codec files. Use concat filter for mixed files.
6. **Scale**: Use `-2` (not `-1`) for auto-dimension to ensure divisibility by 2 (required by H.264).
7. **drawtext**: pass a bundled `fontfile=` — do not rely on system fonts.

### Frontend

1. **No upload step** — work with native file paths from dialogs / `OnFileDrop`, not `FormData`.
2. **Validate extension client-side** before calling an operation; show a clear error for unsupported types.
3. **Event cleanup** — `EventsOff` in `useEffect` cleanup; track the per-event unsubscribe handle.
4. **Responsive** — sidebar collapses on narrow window widths.
5. **Reveal output** — after a job completes, offer "Show in Finder/Explorer" and "Open" actions.

## Testing Strategy

- **Go core**: `go test ./...` — unit-test ffmpeg arg building (`internal/ops`, `internal/ffmpeg/args.go`) and the jobs manager. These need no ffmpeg binary.
- **Integration**: with bundled ffmpeg present, run each operation on a small test clip in a temp dir and assert the output file exists and probes correctly.
- **Frontend**: component tests with Vitest + React Testing Library; mock the generated `wailsjs` bindings.
- **Packaging smoke test**: `wails build` per platform, install the artifact on a clean machine (no system ffmpeg, no dev tools), and run one operation end-to-end.

## Configuration

Desktop apps don't take server env vars. Expose these as in-app settings (persisted to a config file in the OS user config dir, e.g. via `os.UserConfigDir()`):

| Setting | Default | Description |
|---------|---------|-------------|
| `maxConcurrentJobs` | `3` | Parallel ffmpeg processes |
| `defaultOutputDir` | OS Movies/Videos dir | Where the Save dialog opens |
| `tempDir` | `os.TempDir()/videoforge` | Scratch space for intermediates |
| `theme` | `dark` | UI theme (dark/light) |
| `cleanTempOnExit` | `true` | Wipe scratch dir on quit |
```