# VideoForge — Progress & Backlog

> **Living document.** Per `CLAUDE.md`, update this file at the end of every working
> session: move shipped items into "Done", adjust the backlogs, and append a dated
> changelog entry.

**Last updated:** 2026-05-28
**Current version:** v0.1.0 (not yet tagged/released)
**Repo:** github.com/AymanHamdoun/videoforge · default branch `main`

---

## Status snapshot

A working, fully-offline desktop video toolkit built on **Wails v2** (Go core + React/TS
in the native OS webview). All ten planned operations are implemented and the app builds,
bundles ffmpeg, and packages into a macOS `.app`/`.dmg` and a Windows NSIS `.exe` — both
verified building from a macOS machine. No release has been cut yet.

---

## Done

### Core / architecture
- Wails v2 scaffold: Go `App` with bound methods (no HTTP server), React 19 + TS + Vite frontend.
- Job manager (`internal/jobs`): concurrent (semaphore-capped at 3), cancellable, `RunFunc`-based so multi-pass ops fit; progress via Wails events (`job:progress` / `job:done`) — replaces SSE.
- Native file dialogs + drag-and-drop (paths, no uploads). Single + multi-file + image pickers.

### Operations (all 10)
- Convert, Speed, Trim, Compress (single-pass CRF), Extract audio, Merge (concat demuxer), GIF (two-pass palette), Watermark (image overlay), Rotate/flip, Metadata (read-only).
- Arg builders in `internal/ops/*.go`; one bound method each in `app.go`.

### ffmpeg bundling
- `scripts/fetch-ffmpeg.sh` (darwin/linux/windows targets) + `scripts/fetch-ffmpeg.ps1`.
- Auto-bundled: macOS post-build hook → `Contents/Resources/bin`; Windows NSIS → `bin\` beside exe.
- Runtime resolver (`internal/ffmpeg/binaries.go`) with dev fallback to PATH. Binaries git-ignored.

### Build / CI / distribution
- Cross-builds verified from macOS: mac universal `.app` + `.dmg`; Windows `.exe` + NSIS installer (contains `bin/ffmpeg.exe` + `bin/ffprobe.exe`).
- `.github/workflows/release.yml`: tag-triggered (`v*.*.*`) build on native runners → GitHub Release with both artifacts.

### Frontend / UX
- Sidebar nav, Home page (hero + quick-action grid + plain-English glossary), version shown bottom-center of sidebar (`AppVersion()`).
- Shared components: `VideoInput`, `RunBar`, `Field`; `useJob` / `useJobEvents` hooks. Centered content column.

### Tests
- Integration tests (`internal/ops`) run the real bundled ffmpeg: convert round-trip, speed, trim, rotate (asserts dimension swap), audio extract, GIF two-pass.

---

## Tech backlog (engineering)

Ordered roughly by impact/risk.

1. **CI for tests/lint on push & PR** — today CI only runs on tags. Add a `test.yml` that runs `go test`, `go vet`, and `tsc` on every push/PR. *(High value, cheap.)*
2. **Reproducible ffmpeg fetch** — scripts pull "latest"/"release-essentials" → builds aren't reproducible and there's no integrity check. Pin versions + verify checksums. *(Supply-chain.)*
3. **Code signing / notarization** — unsigned artifacts trip macOS Gatekeeper & Windows SmartScreen. Wire Apple Developer ID + notarytool and Windows Authenticode via repo secrets. *(Blocks real distribution.)*
4. **Native arm64 / universal ffmpeg on macOS** — currently x86_64 (Rosetta). Slower, and fails on Apple Silicon Macs without Rosetta. Source an arm64 static build and lipo a universal binary.
5. **Wire Cancel in the UI** — `CancelJob` is bound but no button uses it.
6. **Merge robustness** — stream-copy only; mismatched codec/resolution fails. Add the concat-filter re-encode fallback (REQS §6) with auto-detect.
7. **Settings persistence** — REQS specced app settings (max concurrent jobs, default output dir, theme, temp cleanup) via `os.UserConfigDir`; none implemented. Also persist window size/position.
8. **Unit tests for arg builders + frontend tests** — only integration tests exist. Add table-driven `*_test.go` for `ops` (no ffmpeg needed) and Vitest + RTL for components (specced, not set up).
9. **Compress: target-file-size (two-pass)** — REQS §4 two-pass mode not implemented; only CRF.
10. **Text watermark** — REQS §8 `drawtext`; needs a bundled font for cross-platform consistency.
11. **Friendlier errors** — surface raw ffmpeg stderr tail; map common failures to readable messages.
12. **Temp cleanup / TTL** — per-job temp files are cleaned, but no global scratch-dir sweep on exit (REQS `cleanTempOnExit`).
13. **Accessibility** — drop zones are `div`s with `onClick` (no keyboard focus/aria); add roles, focus states, keyboard handlers.
14. **Auto-update** — no update mechanism; consider a check-for-update against GitHub Releases.
15. **Housekeeping** — remove the commented `replace` directive in `go.mod`; `interface{}` → `any` lint hints; consider a Go-gettable module path.

---

## Product backlog (PM)

1. **Cut v0.1.0** — tag a release to exercise the pipeline and produce the first downloadable installers. The version string is currently hardcoded; tie it to the git tag (ldflags) when releasing.
2. **App icon & branding** — still the default Wails icon. Needed before any public release (Dock/taskbar, installer, dmg).
3. **LICENSE** — no license file. Decide and add (affects whether/how others can use it), especially if the repo is public.
4. **Repo visibility decision** — public = free unlimited CI; private = 2,000 min/mo with macOS billed 10×. Drives CI strategy.
5. **README polish** — add screenshots/GIF, download links, "what is this" for non-technical users.
6. **Real Windows run-test** — Windows builds are *build*-verified only, never *run* on real hardware. Validate the full flow on Windows.
7. **Distribution channels** — beyond GitHub Releases: Homebrew cask, winget, a simple landing page.
8. **First-run / onboarding** — the glossary helps; consider a short "pick a file → choose options → save" walkthrough.
9. **Roadmap & prioritization** — turn this backlog into milestones; decide what's in v0.1 vs later.
10. **Signing cost/budget** — Apple Developer ($99/yr) and a Windows cert are paid; decide if/when.

---

## Suggested next steps (short list)

1. Add a `test.yml` CI workflow (cheap, prevents regressions).
2. Wire the Cancel button + Merge re-encode fallback (most user-visible gaps).
3. Add an app icon + LICENSE, then cut **v0.1.0** to validate the release pipeline end-to-end.

---

## Changelog

- **2026-05-28** — Home page (quick actions + glossary), sidebar version, centered layout fix. Created PROGRESS.MD + CLAUDE.md.
- **2026-05-28** — Implemented the nine remaining operations + sidebar UI; integration tests for new ops.
- **2026-05-27** — Tag-triggered release workflow (mac `.dmg` + win `.exe`). Verified Windows cross-build + NSIS installer from macOS.
- **2026-05-27** — Bundled ffmpeg into the app (fetch scripts + post-build hook + NSIS). Convert wired end-to-end as reference op. Initial Wails scaffold from the desktop-only REQS rewrite.
