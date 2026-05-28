package ops

import "fmt"

// WatermarkParams overlays an image (e.g. logo) onto the video. REQS.MD §8.
type WatermarkParams struct {
	InputPath     string `json:"inputPath"`
	OutputPath    string `json:"outputPath"`
	WatermarkPath string `json:"watermarkPath"`
	Position      string `json:"position"` // top-left|top-right|bottom-left|bottom-right|center
	Margin        int    `json:"margin"`   // px from edges, default 10
}

func WatermarkArgs(p WatermarkParams) []string {
	return []string{
		"-i", p.InputPath,
		"-i", p.WatermarkPath,
		"-filter_complex", fmt.Sprintf("overlay=%s", overlayPosition(p.Position, p.Margin)),
		"-c:v", "libx264", "-crf", "22",
		"-c:a", "copy",
		"-pix_fmt", "yuv420p",
		p.OutputPath,
	}
}

func overlayPosition(pos string, margin int) string {
	if margin == 0 {
		margin = 10
	}
	m := margin
	switch pos {
	case "top-left":
		return fmt.Sprintf("%d:%d", m, m)
	case "top-right":
		return fmt.Sprintf("W-w-%d:%d", m, m)
	case "bottom-left":
		return fmt.Sprintf("%d:H-h-%d", m, m)
	case "center":
		return "(W-w)/2:(H-h)/2"
	case "bottom-right":
		fallthrough
	default:
		return fmt.Sprintf("W-w-%d:H-h-%d", m, m)
	}
}
