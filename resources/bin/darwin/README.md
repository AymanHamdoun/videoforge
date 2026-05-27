# Bundled ffmpeg — macOS

Place static **ffmpeg** and **ffprobe** binaries in this folder:

```
resources/bin/darwin/ffmpeg
resources/bin/darwin/ffprobe
```

These are **not** committed to git (see `.gitignore`) — each is ~40–80 MB. Download
them per-machine / in CI.

## Where to get them

Static universal/arm64 macOS builds:

- https://evermeet.cx/ffmpeg/  (separate `ffmpeg` and `ffprobe` zips)
- or `brew install ffmpeg` and copy the resolved binaries (`$(brew --prefix)/bin/ffmpeg`)

After downloading:

```bash
chmod +x resources/bin/darwin/ffmpeg resources/bin/darwin/ffprobe
xattr -dr com.apple.quarantine resources/bin/darwin/*   # clear Gatekeeper quarantine
```

## How they get into the app

- During `wails dev` the Go core finds them here via the dev-tree fallback in
  `internal/ffmpeg/binaries.go`.
- For a packaged `.app`, copy them into `Contents/Resources/bin/` (post-build hook,
  then re-sign). See README.md → "Packaging".
