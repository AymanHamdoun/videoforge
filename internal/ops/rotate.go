package ops

// RotateParams rotates or flips the video. REQS.MD §9.
type RotateParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Mode       string `json:"mode"` // cw | ccw | 180 | hflip | vflip
}

func RotateArgs(p RotateParams) []string {
	vf := map[string]string{
		"cw":    "transpose=1",
		"ccw":   "transpose=2",
		"180":   "transpose=1,transpose=1",
		"hflip": "hflip",
		"vflip": "vflip",
	}[p.Mode]
	if vf == "" {
		vf = "transpose=1"
	}
	return []string{
		"-i", p.InputPath,
		"-vf", vf,
		"-c:v", "libx264", "-crf", "22",
		"-c:a", "copy",
		"-pix_fmt", "yuv420p",
		p.OutputPath,
	}
}
