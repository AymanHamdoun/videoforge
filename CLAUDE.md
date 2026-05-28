# CLAUDE.md — VideoForge

Guidance for Claude Code working in this repo. Read this first each session.

## ⚠️ Standing instruction: keep PROGRESS.MD current

**At the end of every working session, update [`PROGRESS.MD`](./PROGRESS.MD):**
move shipped work into "Done", revise the Tech/Product backlogs, and append a dated
entry to the Changelog. Treat this as part of finishing any task — not optional.

## What this is

A fully-offline **native desktop** video toolkit. Go core wraps a **bundled**
ffmpeg/ffprobe; React + TypeScript UI runs in the OS-native webview via **Wails v2**.
Ships as a macOS `.dmg` and a Windows NSIS `.exe`. No server, no Docker, no uploads.

- **Spec:** [`REQS.MD`](./REQS.MD) — the design/source-of-truth (ffmpeg recipes, architecture).
- **Status:** [`PROGRESS.MD`](./PROGRESS.MD) — what's done, what's lacking, backlog, changelog.

## Architecture (essentials)

- **No HTTP/router.** The `App` struct's exported methods (`app.go`) are the API; Wails
  binds them to JS (`frontend/wailsjs/go/main/App`). After changing bound methods or
  Go structs, regenerate bindings: `wails generate module`.
- **Progress via events**, not SSE: Go emits `job:progress` / `job:done` (`runtime.EventsEmit`);
  frontend listens via `useJobEvents` / `useJob`.
- **Files are paths**, from native dialogs / `OnFileDrop` — never uploads/multipart.
- **Operations:** pure arg-builders in `internal/ops/<op>.go`; executed by `internal/jobs`
  (concurrent, cancellable, `RunFunc`-based for multi-pass ops like GIF).
- **ffmpeg is bundled** and resolved at runtime by `internal/ffmpeg/binaries.go`
  (app bundle → exe-dir/bin → dev `resources/bin/<os>` → PATH). Always call
  `FFmpegPath()` / `FFprobePath()`, never bare `"ffmpeg"`.
- **Tool list** lives once in `frontend/src/tools/meta.ts` (sidebar + Home grid share it).

## Common commands

```bash
export PATH="$PATH:$(go env GOPATH)/bin"   # so `wails` is on PATH

./scripts/fetch-ffmpeg.sh                  # fetch ffmpeg into resources/bin/<os> (once)
wails dev                                  # hot-reload dev window
wails build -clean                         # build VideoForge.app (ffmpeg auto-bundled)
wails generate module                      # regenerate TS bindings after Go API changes

go build ./... && go vet ./...
go test ./...                              # ops integration tests skip if ffmpeg not fetched
(cd frontend && npm run build)             # tsc + vite typecheck/build
```

## Conventions

- Keep ffmpeg command logic faithful to `REQS.MD`; build args as `[]string` slices,
  no ffmpeg wrapper libraries.
- Cross-platform: use `path/filepath` and `os.DevNull`, never hardcode `/` or `/dev/null`.
- Bundled ffmpeg/ffprobe binaries are **git-ignored** (large) — never commit them. CI/devs
  fetch them. Don't commit `build/bin` or `build/windows/installer/tmp`.
- Match existing code style; the frontend is plain CSS (see `App.css`), dark theme.

## Repo / workflow

- Default branch `main`. Pushes use SSH with the configured key
  (`core.sshCommand` → `~/.ssh/id_personal`); `git push origin main` just works.
- Releases are **tag-triggered**: `git tag vX.Y.Z && git push origin vX.Y.Z` runs
  `.github/workflows/release.yml` (mac + Windows artifacts attached to a GitHub Release).
- Commit messages end with the `Co-Authored-By: Claude` trailer.
