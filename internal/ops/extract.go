package ops

// ExtractParams pulls the audio track out as a standalone file. REQS.MD §5.
type ExtractParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Format     string `json:"format"` // mp3 | wav | flac | aac
}

func ExtractArgs(p ExtractParams) []string {
	base := []string{"-i", p.InputPath, "-vn"} // -vn drops video
	switch p.Format {
	case "mp3":
		return append(base, "-c:a", "libmp3lame", "-q:a", "2", p.OutputPath)
	case "flac":
		return append(base, "-c:a", "flac", p.OutputPath)
	case "aac":
		return append(base, "-c:a", "aac", "-b:a", "256k", p.OutputPath)
	case "wav":
		fallthrough
	default:
		return append(base, p.OutputPath)
	}
}
