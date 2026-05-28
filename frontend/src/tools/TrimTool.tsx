import { useState } from "react";
import { Trim, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { extOf } from "../lib/util";

export function TrimTool() {
  const [input, setInput] = useState("");
  const [start, setStart] = useState("00:00:00");
  const [duration, setDuration] = useState("00:00:10");
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_trimmed", extOf(input)));
      if (!out) return "";
      return Trim({ inputPath: input, outputPath: out, start, duration });
    });

  return (
    <div className="tool">
      <h2>Trim / cut</h2>
      <p className="sub">Fast lossless cut (no re-encode). Times as HH:MM:SS or seconds.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Start">
            <input value={start} onChange={(e) => setStart(e.target.value)} placeholder="00:00:00" />
          </Field>
          <Field label="Duration to keep">
            <input value={duration} onChange={(e) => setDuration(e.target.value)} placeholder="00:00:10" />
          </Field>
          <RunBar job={job} label="Trim" onRun={run} disabled={!input} />
        </div>
      )}
    </div>
  );
}
