package ops

import (
	"fmt"
	"strings"
)

// MergeParams concatenates same-codec files via the concat demuxer. REQS.MD §6.
type MergeParams struct {
	InputPaths []string `json:"inputPaths"`
	OutputPath string   `json:"outputPath"`
}

// MergeFileList renders the concat demuxer's filelist contents. Single quotes in
// paths are escaped per ffmpeg's concat syntax.
func MergeFileList(paths []string) string {
	var b strings.Builder
	for _, p := range paths {
		esc := strings.ReplaceAll(p, "'", `'\''`)
		fmt.Fprintf(&b, "file '%s'\n", esc)
	}
	return b.String()
}

func MergeArgs(filelistPath, outputPath string) []string {
	return []string{
		"-f", "concat", "-safe", "0",
		"-i", filelistPath,
		"-c", "copy",
		outputPath,
	}
}
