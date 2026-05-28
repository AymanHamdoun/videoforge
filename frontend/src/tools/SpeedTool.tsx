import { useState } from "react";
import { ChangeSpeed, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { ToolProps } from "../lib/nav";

const PRESETS = [0.25, 0.5, 1, 1.5, 2, 4];
const MIN = 0.1;
const MAX = 16;

export function SpeedTool({ initialInput }: ToolProps) {
  const [input, setInput] = useState(initialInput ?? "");
  const [speed, setSpeed] = useState(2);
  const job = useJob();

  const valid = speed >= MIN && speed <= MAX;

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
          <Field label={`Speed multiplier (${MIN}× – ${MAX}×)`}>
            <input
              type="number"
              min={MIN}
              max={MAX}
              step={0.05}
              value={speed}
              onChange={(e) => setSpeed(parseFloat(e.target.value))}
            />
          </Field>
          <div className="chips">
            {PRESETS.map((s) => (
              <button
                key={s}
                className={`chip ${speed === s ? "active" : ""}`}
                onClick={() => setSpeed(s)}
              >
                {s}×
              </button>
            ))}
          </div>
          {!valid && (
            <p className="error">Enter a speed between {MIN}× and {MAX}×.</p>
          )}
          <RunBar
            job={job}
            label={valid ? `Make it ${speed}×` : "Make it…"}
            onRun={run}
            disabled={!input || !valid}
            self="speed"
          />
        </div>
      )}
    </div>
  );
}
