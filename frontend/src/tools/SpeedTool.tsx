import { useState } from "react";
import { ChangeSpeed, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";

const SPEEDS = [0.25, 0.5, 0.75, 1, 1.5, 2, 3, 4];

export function SpeedTool() {
  const [input, setInput] = useState("");
  const [speed, setSpeed] = useState(2);
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, `_${speed}x`, ".mp4"));
      if (!out) return "";
      return ChangeSpeed({ inputPath: input, outputPath: out, speed });
    });

  return (
    <div className="tool">
      <h2>Change speed</h2>
      <p className="sub">Speed up or slow down video and audio together.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label={`Speed: ${speed}×`}>
            <select value={speed} onChange={(e) => setSpeed(+e.target.value)}>
              {SPEEDS.map((s) => (
                <option key={s} value={s}>
                  {s}×
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label={`Make it ${speed}×`} onRun={run} disabled={!input} />
        </div>
      )}
    </div>
  );
}
