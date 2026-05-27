package ffmpeg

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// VideoForge ships ffmpeg/ffprobe inside the app bundle. These helpers resolve
// the correct binary path at runtime so we never depend on a system install.
//
// Resolution order (first match wins):
//  1. $VIDEOFORGE_FFMPEG / $VIDEOFORGE_FFPROBE env override (handy for tests)
//  2. macOS app bundle:   <App>.app/Contents/Resources/bin/<name>
//  3. Next to the binary: <execDir>/bin/<name>            (Windows installer layout)
//  4. Dev tree fallback:  ./resources/bin/<goos>/<name>   (during `wails dev`)
//  5. System PATH         (so contributors with a local ffmpeg can still run)
var (
	ffmpegOnce  sync.Once
	ffprobeOnce sync.Once
	ffmpegPath  string
	ffprobePath string
)

func FFmpegPath() string {
	ffmpegOnce.Do(func() { ffmpegPath = resolveBin("ffmpeg", "VIDEOFORGE_FFMPEG") })
	return ffmpegPath
}

func FFprobePath() string {
	ffprobeOnce.Do(func() { ffprobePath = resolveBin("ffprobe", "VIDEOFORGE_FFPROBE") })
	return ffprobePath
}

func resolveBin(name, envKey string) string {
	bin := name
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	if override := os.Getenv(envKey); override != "" {
		return override
	}

	var candidates []string
	if exe, err := os.Executable(); err == nil {
		execDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(execDir, "..", "Resources", "bin", bin), // macOS .app bundle
			filepath.Join(execDir, "bin", bin),                     // alongside the exe
		)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "resources", "bin", runtime.GOOS, bin), // dev tree
		)
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}

	// Last resort: let the OS find it on PATH. Returns the bare name if absent,
	// which makes the eventual exec error message obvious.
	if p, err := exec.LookPath(bin); err == nil {
		return p
	}
	return bin
}
