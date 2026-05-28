package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Thumbnail extracts a single frame as JPEG bytes, scaled to width px wide
// (height auto). atSec is the seek position in seconds.
func Thumbnail(path string, width int, atSec float64) ([]byte, error) {
	if width <= 0 {
		width = 480
	}
	tmp, err := os.CreateTemp("", "vf-thumb-*.jpg")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	args := []string{
		"-y",
		"-ss", strconv.FormatFloat(atSec, 'f', 3, 64),
		"-i", path,
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale=%d:-1", width),
		tmpPath,
	}
	cmd := exec.CommandContext(context.Background(), FFmpegPath(), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("thumbnail: %w: %s", err, out)
	}
	return os.ReadFile(tmpPath)
}
