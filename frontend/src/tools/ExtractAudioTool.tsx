import { useState } from "react";
import { ExtractAudio, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";

const FORMATS = ["mp3", "wav", "flac", "aac"];

export function ExtractAudioTool() {
  const [input, setInput] = useState("");
  const [format, setFormat] = useState("mp3");
  const job = useJob();

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_audio", "." + format));
      if (!out) return "";
      return ExtractAudio({ inputPath: input, outputPath: out, format });
    });

  return (
    <div className="tool">
      <h2>Extract audio</h2>
      <p className="sub">Pull the soundtrack out as a standalone audio file.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Audio format">
            <select value={format} onChange={(e) => setFormat(e.target.value)}>
              {FORMATS.map((f) => (
                <option key={f} value={f}>
                  {f.toUpperCase()}
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label={`Extract ${format.toUpperCase()}`} onRun={run} disabled={!input} />
        </div>
      )}
    </div>
  );
}
