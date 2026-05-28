import { useState } from "react";
import { Rotate, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { extOf } from "../lib/util";

const MODES = [
  { id: "cw", label: "Rotate 90° clockwise" },
  { id: "ccw", label: "Rotate 90° counter-clockwise" },
  { id: "180", label: "Rotate 180°" },
  { id: "hflip", label: "Flip horizontal" },
  { id: "vflip", label: "Flip vertical" },
];

export function RotateTool() {
  const [input, setInput] = useState("");
  const [mode, setMode] = useState("cw");
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_rotated", extOf(input)));
      if (!out) return "";
      return Rotate({ inputPath: input, outputPath: out, mode });
    });

  return (
    <div className="tool">
      <h2>Rotate / flip</h2>
      <p className="sub">Reorient the video.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Transform">
            <select value={mode} onChange={(e) => setMode(e.target.value)}>
              {MODES.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.label}
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label="Apply" onRun={run} disabled={!input} />
        </div>
      )}
    </div>
  );
}
