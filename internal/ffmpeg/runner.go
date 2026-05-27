package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// ProgressCallback receives a 0-100 completion percentage.
type ProgressCallback func(percent float64)

var outTimeRegex = regexp.MustCompile(`out_time_us=(\d+)`)

// Run executes ffmpeg with the given args, parsing -progress output to report
// percentage. The caller's args should NOT include -y/-progress/-nostats; those
// are prepended here. totalDuration is in seconds (0 disables percentage).
func Run(ctx context.Context, args []string, totalDuration float64, onProgress ProgressCallback) error {
	fullArgs := append([]string{"-y", "-progress", "pipe:1", "-nostats"}, args...)
	cmd := exec.CommandContext(ctx, FFmpegPath(), fullArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}
	// Keep stderr for a useful error message on failure.
	var stderr stderrBuf
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if m := outTimeRegex.FindStringSubmatch(scanner.Text()); len(m) == 2 {
			us, _ := strconv.ParseFloat(m[1], 64)
			if totalDuration > 0 && onProgress != nil {
				percent := (us / 1_000_000) / totalDuration * 100
				if percent > 100 {
					percent = 100
				}
				onProgress(percent)
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, stderr.tail())
	}
	if onProgress != nil {
		onProgress(100)
	}
	return nil
}

// stderrBuf keeps only the tail of stderr to surface in error messages.
type stderrBuf struct{ buf []byte }

func (s *stderrBuf) Write(p []byte) (int, error) {
	s.buf = append(s.buf, p...)
	const max = 4096
	if len(s.buf) > max {
		s.buf = s.buf[len(s.buf)-max:]
	}
	return len(p), nil
}

func (s *stderrBuf) tail() string { return string(s.buf) }
