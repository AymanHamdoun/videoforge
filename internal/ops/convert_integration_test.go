package ops_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"videoforge/internal/ffmpeg"
	"videoforge/internal/ops"
)

// repoRoot walks up from this test file to the module root.
func repoRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// TestConvertRoundTrip generates a short clip with the bundled ffmpeg, converts
// it through the real ops + runner path, and asserts the output is valid.
// Skips if the bundled binaries aren't present (e.g. before fetch-ffmpeg.sh).
func TestConvertRoundTrip(t *testing.T) {
	root := repoRoot(t)
	binDir := filepath.Join(root, "resources", "bin", runtime.GOOS)
	ffmpegBin := filepath.Join(binDir, "ffmpeg")
	ffprobeBin := filepath.Join(binDir, "ffprobe")
	if runtime.GOOS == "windows" {
		ffmpegBin += ".exe"
		ffprobeBin += ".exe"
	}
	if _, err := os.Stat(ffmpegBin); err != nil {
		t.Skipf("bundled ffmpeg not found at %s (run scripts/fetch-ffmpeg.sh)", ffmpegBin)
	}

	// Point the resolver at the bundled binaries regardless of cwd.
	t.Setenv("VIDEOFORGE_FFMPEG", ffmpegBin)
	t.Setenv("VIDEOFORGE_FFPROBE", ffprobeBin)

	dir := t.TempDir()
	src := filepath.Join(dir, "src.mov")
	out := filepath.Join(dir, "out.mp4")

	// Generate a 2s 320x240 test clip directly with ffmpeg.
	gen := exec.Command(ffmpegBin, "-y", "-f", "lavfi",
		"-i", "testsrc=duration=2:size=320x240:rate=24", src)
	if b, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generate clip: %v\n%s", err, b)
	}

	// Convert through the actual app code path.
	args := ops.ConvertArgs(ops.ConvertParams{
		InputPath: src, OutputPath: out, CRF: 23, Preset: "ultrafast",
	})
	dur := ffmpeg.Duration(src)
	if dur <= 0 {
		t.Fatalf("probe returned non-positive duration: %v", dur)
	}

	var lastPct float64
	if err := ffmpeg.Run(context.Background(), args, dur, func(p float64) { lastPct = p }); err != nil {
		t.Fatalf("convert: %v", err)
	}

	if lastPct != 100 {
		t.Errorf("expected final progress 100, got %v", lastPct)
	}
	info, err := ffmpeg.Probe(out)
	if err != nil {
		t.Fatalf("probe output: %v", err)
	}
	if len(info.Streams) == 0 {
		t.Fatal("output has no streams")
	}
	hasH264 := false
	for _, s := range info.Streams {
		if s.CodecType == "video" && s.CodecName == "h264" {
			hasH264 = true
		}
	}
	if !hasH264 {
		t.Errorf("expected h264 video stream, got %+v", info.Streams)
	}
}
