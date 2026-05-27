import { useEffect, useState } from "react";
import {
  Convert,
  GetMetadata,
  RevealInFolder,
  SelectInputFile,
  SelectOutputPath,
  SuggestOutputName,
} from "../wailsjs/go/main/App";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";
import { ffmpeg } from "../wailsjs/go/models";
import { useJobEvents } from "./hooks/useJobEvents";
import "./App.css";

function App() {
  const [inputPath, setInputPath] = useState<string>("");
  const [meta, setMeta] = useState<ffmpeg.MediaInfo | null>(null);
  const [crf, setCrf] = useState(23);
  const [preset, setPreset] = useState("medium");
  const [jobId, setJobId] = useState<string | null>(null);
  const [err, setErr] = useState<string>("");

  const job = useJobEvents(jobId);

  // Native drag-and-drop of file paths into the window.
  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      if (paths.length > 0) loadInput(paths[0]);
    }, false);
    return () => OnFileDropOff();
  }, []);

  async function loadInput(path: string) {
    setErr("");
    setInputPath(path);
    setJobId(null);
    try {
      setMeta(await GetMetadata(path));
    } catch (e) {
      setMeta(null);
      setErr(String(e));
    }
  }

  async function pickInput() {
    const path = await SelectInputFile();
    if (path) loadInput(path);
  }

  async function runConvert() {
    setErr("");
    try {
      const suggested = await SuggestOutputName(inputPath, ".mp4");
      const outputPath = await SelectOutputPath(suggested);
      if (!outputPath) return; // user canceled save dialog
      const id = await Convert({ inputPath, outputPath, crf, preset });
      setJobId(id);
    } catch (e) {
      setErr(String(e));
    }
  }

  const videoStream = meta?.streams.find((s) => s.codec_type === "video");
  const durationSec = meta ? parseFloat(meta.format.duration || "0") : 0;

  return (
    <div className="app">
      <header className="topbar">
        <span className="logo">🎬 VideoForge</span>
        <span className="tag">Format Converter</span>
      </header>

      <main className="content">
        <section
          className={`drop ${inputPath ? "has-file" : ""}`}
          onClick={pickInput}
        >
          {inputPath ? (
            <div className="fileinfo">
              <div className="filename">{inputPath.split(/[\\/]/).pop()}</div>
              {videoStream && (
                <div className="meta">
                  {videoStream.codec_name} · {videoStream.width}×{videoStream.height} ·{" "}
                  {durationSec.toFixed(1)}s
                </div>
              )}
            </div>
          ) : (
            <div className="placeholder">
              <strong>Drop a video here</strong>
              <span>or click to browse</span>
            </div>
          )}
        </section>

        {inputPath && (
          <section className="controls">
            <label>
              Quality (CRF {crf})
              <input
                type="range"
                min={18}
                max={32}
                value={crf}
                onChange={(e) => setCrf(Number(e.target.value))}
              />
            </label>
            <label>
              Preset
              <select value={preset} onChange={(e) => setPreset(e.target.value)}>
                {["ultrafast", "veryfast", "fast", "medium", "slow", "veryslow"].map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))}
              </select>
            </label>
            <button
              className="primary"
              onClick={runConvert}
              disabled={job.status === "processing"}
            >
              {job.status === "processing" ? "Converting…" : "Convert to MP4"}
            </button>
          </section>
        )}

        {job.status !== "idle" && (
          <section className="progress">
            <div className="bar">
              <div className="fill" style={{ width: `${job.percent}%` }} />
            </div>
            <div className="status">
              {job.status === "processing" && `${job.percent.toFixed(0)}%`}
              {job.status === "completed" && (
                <>
                  ✅ Done —{" "}
                  <button className="link" onClick={() => RevealInFolder(job.output!)}>
                    Show in folder
                  </button>
                </>
              )}
              {job.status === "failed" && <span className="error">❌ {job.error}</span>}
              {job.status === "canceled" && "Canceled"}
            </div>
          </section>
        )}

        {err && <p className="error">{err}</p>}
      </main>
    </div>
  );
}

export default App;
