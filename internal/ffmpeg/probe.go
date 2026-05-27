package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// MediaInfo is the subset of ffprobe JSON output we expose to the UI.
type MediaInfo struct {
	Format  Format   `json:"format"`
	Streams []Stream `json:"streams"`
}

type Format struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"` // seconds, as a string
	Size       string `json:"size"`     // bytes, as a string
	BitRate    string `json:"bit_rate"`
}

type Stream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"` // "video" | "audio" | ...
	CodecName string `json:"codec_name"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	RFrameRate string `json:"r_frame_rate,omitempty"`
}

// Probe returns metadata for the given media file.
func Probe(path string) (*MediaInfo, error) {
	cmd := exec.Command(FFprobePath(),
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}
	var info MediaInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("ffprobe decode: %w", err)
	}
	return &info, nil
}

// Duration returns the media duration in seconds (0 if unknown). Used to drive
// progress percentage before kicking off a job.
func Duration(path string) float64 {
	info, err := Probe(path)
	if err != nil {
		return 0
	}
	d, _ := strconv.ParseFloat(info.Format.Duration, 64)
	return d
}
