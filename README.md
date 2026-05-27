# VideoForge

A fully offline, **native desktop** video toolkit. A Go core wraps a bundled
ffmpeg/ffprobe; a React + TypeScript frontend runs in the OS-native webview via
[Wails](https://wails.io) — no Chromium, no Docker, no server. Ships as a `.dmg`
(macOS) / `.exe` installer (Windows).

See [`REQS.MD`](./REQS.MD) for the full spec.

## Status

Scaffold with the **Format Converter** wired end-to-end as the reference operation:
native file pick / drag-and-drop → ffprobe metadata → ffmpeg convert → live progress
via Wails events → "Show in folder". The remaining operations (speed, trim, compress,
extract, merge, gif, watermark, rotate) follow the same pattern — add `internal/ops/<op>.go`
+ a bound method on `App` + a UI section.

## Architecture

| Concern | Implementation |
|---------|----------------|
| Shell | Wails v2 (native webview) |
| Core API | Exported methods on `App` (`app.go`), bound to JS — no HTTP/router |
| Progress | `runtime.EventsEmit("job:progress"/"job:done")` → `useJobEvents` hook (no SSE) |
| Files | Native open/save dialogs + `OnFileDrop` — work with paths, no uploads |
| ffmpeg | Bundled binaries resolved by `internal/ffmpeg/binaries.go` |
| Jobs | `internal/jobs` — in-memory, semaphore-capped, cancellable |

```
main.go                    Wails entry (embed frontend, bind App, enable file drop)
app.go                     Bound methods = the API
internal/ffmpeg/           runner (progress), probe, binaries resolver
internal/jobs/             concurrent job manager + event emitter
internal/ops/              one file per operation (arg builders)
frontend/src/              React UI; wailsjs/ holds generated bindings
resources/bin/<os>/        bundled ffmpeg/ffprobe (downloaded, not committed)
```

## Prerequisites

- Go 1.23+
- Node 18+ / npm
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## ffmpeg is bundled (no separate install for end users)

The shipped app is fully self-contained — `ffmpeg`/`ffprobe` travel inside it, so a
user just installs the app and everything works offline. As a developer you fetch the
static binaries once into `resources/bin/<os>/`:

```bash
./scripts/fetch-ffmpeg.sh                                  # macOS / Linux
powershell -ExecutionPolicy Bypass -File scripts\fetch-ffmpeg.ps1   # Windows
```

The build then bundles them automatically (macOS post-build hook → `Contents/Resources/bin`;
Windows NSIS installer → `bin\` next to the exe). At runtime `internal/ffmpeg/binaries.go`
resolves the bundled binary, falling back to a system `ffmpeg` on PATH for dev convenience.

## Develop

```bash
wails dev      # native window with hot-reloaded frontend
```

(If you skip the fetch step, dev still works as long as ffmpeg is on your PATH.)

## Build & Package

### macOS — `.app` (ffmpeg auto-bundled) then `.dmg`

```bash
./scripts/fetch-ffmpeg.sh                       # once
wails build -platform darwin/universal -clean
# → build/bin/VideoForge.app  (ffmpeg already inside Contents/Resources/bin)

# For distribution: codesign --deep --sign "Developer ID Application: ..." the .app
# (signs the bundled binaries too), then notarize with `xcrun notarytool`.
# For local use, right-click → Open to bypass Gatekeeper.
create-dmg "VideoForge.dmg" "build/bin/VideoForge.app"     # brew install create-dmg
```

### Windows — NSIS installer `.exe` (ffmpeg auto-bundled)

```powershell
powershell -ExecutionPolicy Bypass -File scripts\fetch-ffmpeg.ps1   # once
wails build -platform windows/amd64 -nsis -clean
# → build/bin/VideoForge-amd64-installer.exe
```

The installer copies `ffmpeg.exe`/`ffprobe.exe` into `bin\` beside `VideoForge.exe` and
already bundles the WebView2 runtime bootstrapper (`wails.webview2runtime` macro) for
older Windows 10 machines.

> **Note on macOS arch:** `scripts/fetch-ffmpeg.sh` pulls an x86_64 static build
> (evermeet.cx) which runs natively on Intel and under Rosetta 2 on Apple Silicon. For a
> native arm64 / universal ffmpeg, point the script at an arm64 static build — the bundling
> layout is identical.
