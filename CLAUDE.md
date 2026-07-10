# CLAUDE.md — VideoForge

Guidance for Claude Code working in this repo. Read this first each session.

## ⚠️ Standing instruction: keep PROGRESS.MD current

**At the end of every working session, update [`PROGRESS.MD`](./PROGRESS.MD):**
move shipped work into "Done", revise the Tech/Product backlogs, and append a dated
entry to the Changelog. Treat this as part of finishing any task — not optional.

## What this is

A **native desktop** video toolkit. Go core wraps a **bundled** ffmpeg/ffprobe;
React + TypeScript UI runs in the OS-native webview via **Wails v2**. Ships as a
macOS `.dmg` and a Windows NSIS `.exe`. No file uploads.

Licensing is handled by **Lemon Squeezy**: the app calls LS's public License API
to activate / validate / deactivate. Once activated, it works offline for up to
**14 days** before LS must be reachable again — see `internal/license`.

- **Spec:** [`REQS.MD`](./REQS.MD) — the design/source-of-truth (ffmpeg recipes, architecture).
- **Status:** [`PROGRESS.MD`](./PROGRESS.MD) — what's done, what's lacking, backlog, changelog.
- **System overview:** [`../README.md`](../README.md) — how the desktop fits with `videoforge-web`.

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
- **Licensing** is a thin HTTP client over Lemon Squeezy's public License API
  (`internal/license/license.go`). Activation hits `POST /v1/licenses/activate`
  with `instance_name = "VideoForge on <hostname>"`; `LoadStatus` revalidates on
  each launch with a 10s timeout, falling back to cached state within the
  14-day grace if LS is unreachable. Permanent rejections (key disabled,
  instance gone) clear the cache and force re-activation. Cached state lives at
  `<UserConfigDir>/VideoForge/license.json`. **Never reintroduce own-signing;
  LS is the source of truth.**

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
- Releases are **cut locally** from a Mac, not by CI: `git tag vX.Y.Z` then
  `./scripts/release.sh` builds the mac `.dmg` + Windows NSIS `.exe` and publishes
  them as **GitHub Release** assets via the `gh` CLI (`brew install gh && gh auth
  login`). The web app's `/api/download/[platform]` redirects to these assets
  (its `GITHUB_RELEASES_REPO` points here). There is no CI pipeline for releases —
  running on a Mac is cheaper than paying for SaaS macOS runners. (GitHub Actions
  CI in `.github/workflows/ci.yml` covers frontend + Go-core build/vet/test only.)
- Commit messages end with the `Co-Authored-By: Claude` trailer.
