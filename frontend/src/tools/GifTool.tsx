import { useState } from "react";
import { MakeGif, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { ToolProps } from "../lib/nav";

export function GifTool({ initialInput }: ToolProps) {
  const [input, setInput] = useState(initialInput ?? "");
  const [start, setStart] = useState("00:00:00");
  const [duration, setDuration] = useState("00:00:05");
  const [fps, setFps] = useState(12);
  const [width, setWidth] = useState(480);
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "", ".gif"));
      if (!out) return "";
      return MakeGif({ inputPath: input, outputPath: out, start, duration, fps, width });
    });

  return (
    <div className="tool">
      <h2>Make GIF</h2>
      <p className="sub">High-quality animated GIF (two-pass palette).</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Start">
            <input value={start} onChange={(e) => setStart(e.target.value)} placeholder="00:00:00" />
          </Field>
          <Field label="Duration">
            <input value={duration} onChange={(e) => setDuration(e.target.value)} placeholder="00:00:05" />
          </Field>
          <Field label={`Frame rate: ${fps} fps`}>
            <input type="range" min={5} max={30} value={fps} onChange={(e) => setFps(+e.target.value)} />
          </Field>
          <Field label={`Width: ${width}px`}>
            <input type="range" min={120} max={960} step={20} value={width} onChange={(e) => setWidth(+e.target.value)} />
          </Field>
          <RunBar job={job} label="Create GIF" onRun={run} disabled={!input} self="gif" />
        </div>
      )}
    </div>
  );
}
