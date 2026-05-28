package ops

import "fmt"

// GifParams makes a high-quality GIF via two-pass palettegen/paletteuse. REQS.MD §7.
type GifParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Start      string `json:"start"`    // optional seek
	Duration   string `json:"duration"` // optional length
	FPS        int    `json:"fps"`      // default 12
	Width      int    `json:"width"`    // default 480, height auto
}

func (p GifParams) fps() int {
	if p.FPS == 0 {
		return 12
	}
	return p.FPS
}

func (p GifParams) width() int {
	if p.Width == 0 {
		return 480
	}
	return p.Width
}

// GifPaletteArgs (pass 1) generates an optimal palette to palettePath.
func GifPaletteArgs(p GifParams, palettePath string) []string {
	args := []string{}
	if p.Start != "" {
		args = append(args, "-ss", p.Start)
	}
	args = append(args, "-i", p.InputPath)
	if p.Duration != "" {
		args = append(args, "-t", p.Duration)
	}
	vf := fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos,palettegen", p.fps(), p.width())
	return append(args, "-vf", vf, palettePath)
}

// GifArgs (pass 2) renders the GIF using the generated palette.
func GifArgs(p GifParams, palettePath string) []string {
	args := []string{}
	if p.Start != "" {
		args = append(args, "-ss", p.Start)
	}
	args = append(args, "-i", p.InputPath, "-i", palettePath)
	if p.Duration != "" {
		args = append(args, "-t", p.Duration)
	}
	fc := fmt.Sprintf("[0:v]fps=%d,scale=%d:-1:flags=lanczos[v];[v][1:v]paletteuse", p.fps(), p.width())
	return append(args, "-filter_complex", fc, p.OutputPath)
}
