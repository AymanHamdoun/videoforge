package ops

import (
	"fmt"
	"strings"
)

// SpeedParams changes playback speed of video + audio in sync. REQS.MD §2.
type SpeedParams struct {
	InputPath  string  `json:"inputPath"`
	OutputPath string  `json:"outputPath"`
	Speed      float64 `json:"speed"` // 0.25 .. 4.0 (1.0 = unchanged)
}

func SpeedArgs(p SpeedParams) []string {
	videoPTS := fmt.Sprintf("setpts=%f*PTS", 1.0/p.Speed)
	atempo := buildAtempoChain(p.Speed)
	return []string{
		"-i", p.InputPath,
		"-filter_complex", fmt.Sprintf("[0:v]%s[v];[0:a]%s[a]", videoPTS, atempo),
		"-map", "[v]", "-map", "[a]",
		"-c:v", "libx264", "-preset", "fast", "-crf", "22",
		"-c:a", "aac", "-b:a", "192k",
		"-pix_fmt", "yuv420p",
		p.OutputPath,
	}
}

// buildAtempoChain composes atempo filters since each is limited to [0.5, 2.0].
func buildAtempoChain(speed float64) string {
	var filters []string
	remaining := speed
	for remaining > 2.0 {
		filters = append(filters, "atempo=2.0")
		remaining /= 2.0
	}
	for remaining < 0.5 {
		filters = append(filters, "atempo=0.5")
		remaining /= 0.5
	}
	filters = append(filters, fmt.Sprintf("atempo=%f", remaining))
	return strings.Join(filters, ",")
}
