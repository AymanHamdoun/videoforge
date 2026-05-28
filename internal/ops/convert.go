package ops

import (
	"path/filepath"
	"strconv"
	"strings"
)

// ConvertParams describes a format-conversion request from the UI.
type ConvertParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	// Optional explicit encode controls. When CRF is 0 and Preset is empty,
	// ffmpeg auto-selects codecs from the output extension.
	Preset string `json:"preset,omitempty"` // ultrafast..veryslow
	CRF    int    `json:"crf,omitempty"`    // 0-51 (0 = let ffmpeg decide)
}

// ConvertArgs builds the ffmpeg argument slice (without the -y/-progress flags,
// which ffmpeg.Run prepends). Codecs are chosen from the OUTPUT container:
// WebM requires VP9/Opus (it can't hold H.264/AAC), everything else uses
// H.264/AAC. Mirrors REQS.MD §1 "Format Conversion".
func ConvertArgs(p ConvertParams) []string {
	args := []string{"-i", p.InputPath}

	// WebM only supports VP8/VP9/AV1 video + Vorbis/Opus audio.
	if strings.ToLower(filepath.Ext(p.OutputPath)) == ".webm" {
		crf := p.CRF
		if crf == 0 {
			crf = 31 // sane VP9 default (its CRF scale is 0-63)
		}
		return append(args,
			"-c:v", "libvpx-vp9",
			"-crf", strconv.Itoa(crf),
			"-b:v", "0", // required for VP9 constant-quality (CRF) mode
			"-row-mt", "1", "-deadline", "good", "-cpu-used", "4", // keep it tolerably fast
			"-c:a", "libopus",
			p.OutputPath,
		)
	}

	if p.Preset != "" || p.CRF > 0 {
		preset := p.Preset
		if preset == "" {
			preset = "medium"
		}
		crf := p.CRF
		if crf == 0 {
			crf = 23
		}
		args = append(args,
			"-c:v", "libx264",
			"-preset", preset,
			"-crf", strconv.Itoa(crf),
			"-c:a", "aac",
			"-b:a", "192k",
			"-pix_fmt", "yuv420p", // CRITICAL: broad player compatibility
		)
	}

	return append(args, p.OutputPath)
}
