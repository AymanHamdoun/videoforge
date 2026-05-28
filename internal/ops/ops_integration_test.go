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

// bundledFFmpeg returns the bundled ffmpeg path, or skips if not fetched yet.
func bundledFFmpeg(t *testing.T) (ffmpegBin string) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	binDir := filepath.Join(root, "resources", "bin", runtime.GOOS)
	ffmpegBin = filepath.Join(binDir, "ffmpeg")
	ffprobeBin := filepath.Join(binDir, "ffprobe")
	if runtime.GOOS == "windows" {
		ffmpegBin += ".exe"
		ffprobeBin += ".exe"
	}
	if _, err := os.Stat(ffmpegBin); err != nil {
		t.Skipf("bundled ffmpeg not found (run scripts/fetch-ffmpeg.sh)")
	}
	t.Setenv("VIDEOFORGE_FFMPEG", ffmpegBin)
	t.Setenv("VIDEOFORGE_FFPROBE", ffprobeBin)
	return ffmpegBin
}

// makeClip generates a 2s clip with both a video and an audio stream.
func makeClip(t *testing.T, ffmpegBin, path string) {
	t.Helper()
	cmd := exec.Command(ffmpegBin, "-y",
		"-f", "lavfi", "-i", "testsrc=duration=2:size=320x240:rate=24",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-shortest", path)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate clip: %v\n%s", err, b)
	}
}

func runAndProbe(t *testing.T, args []string, dur float64, out string) *ffmpeg.MediaInfo {
	t.Helper()
	if err := ffmpeg.Run(context.Background(), args, dur, nil); err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	info, err := ffmpeg.Probe(out)
	if err != nil {
		t.Fatalf("probe %s: %v", out, err)
	}
	return info
}

func hasStream(info *ffmpeg.MediaInfo, kind string) bool {
	for _, s := range info.Streams {
		if s.CodecType == kind {
			return true
		}
	}
	return false
}

func TestOperations(t *testing.T) {
	ff := bundledFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.mp4")
	makeClip(t, ff, src)

	t.Run("speed", func(t *testing.T) {
		out := filepath.Join(dir, "speed.mp4")
		info := runAndProbe(t, ops.SpeedArgs(ops.SpeedParams{InputPath: src, OutputPath: out, Speed: 2}), 1, out)
		if !hasStream(info, "video") || !hasStream(info, "audio") {
			t.Errorf("speed output missing video/audio: %+v", info.Streams)
		}
	})

	t.Run("trim", func(t *testing.T) {
		out := filepath.Join(dir, "trim.mp4")
		info := runAndProbe(t, ops.TrimArgs(ops.TrimParams{InputPath: src, OutputPath: out, Start: "0", Duration: "1"}), 1, out)
		if info.Format.Duration == "" {
			t.Errorf("trim produced no duration")
		}
	})

	t.Run("rotate", func(t *testing.T) {
		out := filepath.Join(dir, "rot.mp4")
		info := runAndProbe(t, ops.RotateArgs(ops.RotateParams{InputPath: src, OutputPath: out, Mode: "cw"}), 2, out)
		// 320x240 rotated 90° -> 240x320
		var v *ffmpeg.Stream
		for i := range info.Streams {
			if info.Streams[i].CodecType == "video" {
				v = &info.Streams[i]
			}
		}
		if v == nil || v.Width != 240 || v.Height != 320 {
			t.Errorf("expected 240x320 after rotate, got %+v", v)
		}
	})

	t.Run("extract-mp3", func(t *testing.T) {
		out := filepath.Join(dir, "audio.mp3")
		info := runAndProbe(t, ops.ExtractArgs(ops.ExtractParams{InputPath: src, OutputPath: out, Format: "mp3"}), 2, out)
		if hasStream(info, "video") || !hasStream(info, "audio") {
			t.Errorf("extract should have audio only: %+v", info.Streams)
		}
	})

	t.Run("gif-two-pass", func(t *testing.T) {
		out := filepath.Join(dir, "out.gif")
		palette := filepath.Join(dir, "pal.png")
		p := ops.GifParams{InputPath: src, OutputPath: out, Duration: "1", FPS: 10, Width: 240}
		if err := ffmpeg.Run(context.Background(), ops.GifPaletteArgs(p, palette), 0, nil); err != nil {
			t.Fatalf("palette: %v", err)
		}
		info := runAndProbe(t, ops.GifArgs(p, palette), 1, out)
		if info.Format.FormatName == "" {
			t.Errorf("gif has no format")
		}
	})
}
