import { useState } from "react";
import {
  AddWatermark,
  SelectImageFile,
  SelectOutputPath,
  SuggestOutput,
} from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { baseName, extOf } from "../lib/util";

const POSITIONS = ["bottom-right", "bottom-left", "top-right", "top-left", "center"];

export function WatermarkTool() {
  const [input, setInput] = useState("");
  const [watermark, setWatermark] = useState("");
  const [position, setPosition] = useState("bottom-right");
  const job = useJob();

  async function pickImage() {
    const p = await SelectImageFile();
    if (p) setWatermark(p);
  }

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_watermarked", extOf(input)));
      if (!out) return "";
      return AddWatermark({ inputPath: input, outputPath: out, watermarkPath: watermark, position, margin: 10 });
    });

  return (
    <div className="tool">
      <h2>Add watermark</h2>
      <p className="sub">Overlay a logo/image (PNG with transparency works best).</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />
      {input && (
        <div className="controls">
          <Field label="Watermark image">
            <button className="secondary" onClick={pickImage}>
              {watermark ? baseName(watermark) : "Choose image…"}
            </button>
          </Field>
          <Field label="Position">
            <select value={position} onChange={(e) => setPosition(e.target.value)}>
              {POSITIONS.map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
          </Field>
          <RunBar job={job} label="Add watermark" onRun={run} disabled={!input || !watermark} />
        </div>
      )}
    </div>
  );
}
