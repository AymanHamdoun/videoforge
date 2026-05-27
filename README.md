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
- ffmpeg/ffprobe placed in `resources/bin/<os>/` — see the READMEs there.

## Develop

```bash
wails dev      # native window with hot-reloaded frontend
```

(Without bundled ffmpeg present, the resolver falls back to a system ffmpeg on PATH,
so a local `brew install ffmpeg` is enough for dev.)

## Build & Package

### macOS — `.app` then `.dmg`

```bash
wails build -platform darwin/universal -clean
# → build/bin/VideoForge.app

# Bundle ffmpeg into the .app, then create the dmg:
cp resources/bin/darwin/ffmpeg resources/bin/darwin/ffprobe \
   build/bin/VideoForge.app/Contents/Resources/bin/        # (mkdir bin first)
# For distribution: codesign --deep --sign "Developer ID Application: ..." the .app,
# then notarize with `xcrun notarytool`. For local use, right-click → Open.
create-dmg "VideoForge.dmg" "build/bin/VideoForge.app"     # brew install create-dmg
```

### Windows — NSIS installer `.exe`

```bash
wails build -platform windows/amd64 -nsis -clean
# → build/bin/VideoForge-amd64-installer.exe
```

The installer must also copy `resources/bin/windows/ffmpeg.exe` + `ffprobe.exe` into a
`bin\` folder beside `VideoForge.exe`, and bundle the WebView2 bootstrapper for older
Windows 10 machines (Wails `webview2` install strategy).
