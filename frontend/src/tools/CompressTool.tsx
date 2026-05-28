import { useState } from "react";
import { Compress, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { ToolProps } from "../lib/nav";

export function CompressTool({ initialInput }: ToolProps) {
  const [input, setInput] = useState(initialInput ?? "");
  const [crf, setCrf] = useState(28);
  const [preset, setPreset] = useState("slow");
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_compressed", ".mp4"));
      if (!out) return "";
      return Compress({ inputPath: input, outputPath: out, crf, preset });
    });

  return (
    <div className="tool">
      <h2>Compress</h2>
      <p className="sub">Shrink file size with H.264. Higher CRF = smaller file, lower quality.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label={`Compression (CRF ${crf})`}>
            <input type="range" min={23} max={40} value={crf} onChange={(e) => setCrf(+e.target.value)} />
          </Field>
          <Field label="Preset (slower = smaller)">
            <select value={preset} onChange={(e) => setPreset(e.target.value)}>
              {["veryfast", "fast", "medium", "slow", "veryslow"].map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label="Compress" onRun={run} disabled={!input} self="compress" />
        </div>
      )}
    </div>
  );
}
