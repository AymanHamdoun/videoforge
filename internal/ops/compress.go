package ops

import "strconv"

// CompressParams reduces file size via CRF. REQS.MD §4 (single-pass mode).
type CompressParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	CRF        int    `json:"crf"`    // 18-51; higher = smaller. Default 28.
	Preset     string `json:"preset"` // default "slow" for better ratio
}

func CompressArgs(p CompressParams) []string {
	crf := p.CRF
	if crf == 0 {
		crf = 28
	}
	preset := p.Preset
	if preset == "" {
		preset = "slow"
	}
	return []string{
		"-i", p.InputPath,
		"-c:v", "libx264",
		"-preset", preset,
		"-crf", strconv.Itoa(crf),
		"-c:a", "aac", "-b:a", "128k",
		"-pix_fmt", "yuv420p",
		p.OutputPath,
	}
}
