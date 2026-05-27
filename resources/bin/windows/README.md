# Bundled ffmpeg — Windows

Place static **ffmpeg.exe** and **ffprobe.exe** in this folder:

```
resources/bin/windows/ffmpeg.exe
resources/bin/windows/ffprobe.exe
```

These are **not** committed to git (see `.gitignore`) — each is large. Download them
per-machine / in CI.

## Where to get them

Static Windows builds:

- https://www.gyan.dev/ffmpeg/builds/  ("release essentials" contains ffmpeg + ffprobe)
- https://github.com/BtbN/FFmpeg-Builds/releases

Extract `ffmpeg.exe` and `ffprobe.exe` from the archive's `bin/` into this folder.

## How they get into the app

- During `wails dev` the Go core finds them here via the dev-tree fallback in
  `internal/ffmpeg/binaries.go`.
- The NSIS installer copies them next to `VideoForge.exe` in a `bin\` subfolder, where
  the production resolver looks. See README.md → "Packaging".
