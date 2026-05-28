import { useState } from "react";
import { Convert, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";

const FORMATS = ["mp4", "mkv", "mov", "webm", "avi"];

export function ConvertTool() {
  const [input, setInput] = useState("");
  const [format, setFormat] = useState("mp4");
  const [crf, setCrf] = useState(23);
  const [preset, setPreset] = useState("medium");
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_converted", "." + format));
      if (!out) return "";
      return Convert({ inputPath: input, outputPath: out, crf, preset });
    });

  return (
    <div className="tool">
      <h2>Convert format</h2>
      <p className="sub">Re-encode to another container/codec.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Output format">
            <select value={format} onChange={(e) => setFormat(e.target.value)}>
              {FORMATS.map((f) => (
                <option key={f} value={f}>
                  {f.toUpperCase()}
                </option>
              ))}
            </select>
          </Field>
          <Field label={`Quality (CRF ${crf}) — lower is better`}>
            <input type="range" min={18} max={32} value={crf} onChange={(e) => setCrf(+e.target.value)} />
          </Field>
          <Field label="Encoding preset">
            <select value={preset} onChange={(e) => setPreset(e.target.value)}>
              {["ultrafast", "veryfast", "fast", "medium", "slow", "veryslow"].map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label={`Convert to ${format.toUpperCase()}`} onRun={run} disabled={!input} />
        </div>
      )}
    </div>
  );
}
