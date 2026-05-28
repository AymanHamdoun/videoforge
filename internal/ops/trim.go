package ops

// TrimParams cuts a section using fast stream-copy (no re-encode). REQS.MD §3.
// Start/Duration accept "SS", "MM:SS", or "HH:MM:SS".
type TrimParams struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Start      string `json:"start"`    // seek point (-ss before -i = fast)
	Duration   string `json:"duration"` // length to keep
}

func TrimArgs(p TrimParams) []string {
	args := []string{}
	if p.Start != "" {
		args = append(args, "-ss", p.Start)
	}
	args = append(args, "-i", p.InputPath)
	if p.Duration != "" {
		args = append(args, "-t", p.Duration)
	}
	return append(args, "-c", "copy", p.OutputPath)
}
