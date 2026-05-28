import { useEffect, useRef, useState } from "react";
import { Trim, SelectOutputPath, SuggestOutput } from "../../wailsjs/go/main/App";
import { VideoInput } from "../components/VideoInput";
import { RunBar } from "../components/RunBar";
import { Field } from "../components/Field";
import { useJob } from "../hooks/useJob";
import { ToolProps } from "../lib/nav";
import { extOf } from "../lib/util";

function fmt(s: number): string {
  if (!isFinite(s) || s < 0) s = 0;
  const m = Math.floor(s / 60);
  const sec = (s % 60).toFixed(1).padStart(4, "0");
  return `${m}:${sec}`;
}

export function TrimTool({ initialInput }: ToolProps) {
  const [input, setInput] = useState(initialInput ?? "");
  const job = useJob();

  const videoRef = useRef<HTMLVideoElement>(null);
  const trackRef = useRef<HTMLDivElement>(null);
  const [duration, setDuration] = useState(0);
  const [current, setCurrent] = useState(0);
  const [start, setStart] = useState(0);
  const [end, setEnd] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [drag, setDrag] = useState<null | "start" | "end" | "seek">(null);

  // Fallback when the OS webview can't decode this format (e.g. VP9/MKV on macOS).
  const [previewFailed, setPreviewFailed] = useState(false);
  const [mStart, setMStart] = useState("00:00:00");
  const [mDuration, setMDuration] = useState("00:00:10");

  const mediaUrl = input ? `/media?p=${encodeURIComponent(input)}` : "";

  useEffect(() => {
    setDuration(0);
    setCurrent(0);
    setStart(0);
    setEnd(0);
    setPreviewFailed(false);
  }, [input]);

  function seek(t: number) {
    const v = videoRef.current;
    if (v) v.currentTime = t;
    setCurrent(t);
  }

  function onLoaded() {
    const v = videoRef.current;
    if (!v) return;
    setDuration(v.duration);
    setStart(0);
    setEnd(v.duration);
  }

  function togglePlay() {
    const v = videoRef.current;
    if (!v) return;
    if (playing) v.pause();
    else {
      if (v.currentTime < start || v.currentTime >= end) v.currentTime = start;
      v.play();
    }
  }

  function onTimeUpdate() {
    const v = videoRef.current;
    if (!v) return;
    if (v.currentTime >= end) v.currentTime = start; // loop the selection
    setCurrent(v.currentTime);
  }

  useEffect(() => {
    if (!drag) return;
    function move(e: PointerEvent) {
      const track = trackRef.current;
      if (!track || duration <= 0) return;
      const rect = track.getBoundingClientRect();
      const pct = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
      const t = pct * duration;
      if (drag === "start") {
        const ns = Math.max(0, Math.min(t, end - 0.1));
        setStart(ns);
        seek(ns);
      } else if (drag === "end") {
        const ne = Math.min(duration, Math.max(t, start + 0.1));
        setEnd(ne);
        seek(ne);
      } else {
        seek(t);
      }
    }
    function up() {
      setDrag(null);
    }
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
    return () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
  }, [drag, duration, start, end]);

  const pct = (t: number) => (duration > 0 ? (t / duration) * 100 : 0);

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(input, "_trimmed", extOf(input)));
      if (!out) return "";
      const params = previewFailed
        ? { inputPath: input, outputPath: out, start: mStart, duration: mDuration }
        : { inputPath: input, outputPath: out, start: start.toFixed(3), duration: (end - start).toFixed(3) };
      return Trim(params);
    });

  const canRun = previewFailed ? !!mDuration.trim() : end - start >= 0.1;

  return (
    <div className="tool">
      <h2>Trim / cut</h2>
      <p className="sub">Scrub the video and drag the handles to set the start and end.</p>
      <VideoInput path={input} onChange={(p) => setInput(p)} />

      {input && (
        <div className="trimmer">
          {/* Always mount the video so it can attempt to load / report errors,
              but hide it when the format isn't previewable. */}
          <video
            ref={videoRef}
            src={mediaUrl}
            className="trim-video"
            style={previewFailed ? { display: "none" } : undefined}
            onLoadedMetadata={onLoaded}
            onTimeUpdate={onTimeUpdate}
            onError={() => setPreviewFailed(true)}
            onPlay={() => setPlaying(true)}
            onPause={() => setPlaying(false)}
            onClick={togglePlay}
          />

          {previewFailed ? (
            <div className="preview-fallback">
              <p className="notice">
                ⚠️ Live preview isn't available for this format on this OS (common for
                <code> .webm</code> / <code>.mkv</code> on macOS). Trimming still works —
                just enter the times below. Tip: use the <strong>Metadata</strong> tool to
                check the total duration.
              </p>
              <Field label="Start (HH:MM:SS or seconds)">
                <input value={mStart} onChange={(e) => setMStart(e.target.value)} placeholder="00:00:00" />
              </Field>
              <Field label="Duration to keep (HH:MM:SS or seconds)">
                <input value={mDuration} onChange={(e) => setMDuration(e.target.value)} placeholder="00:00:10" />
              </Field>
            </div>
          ) : (
            <>
              <div
                className="timeline"
                ref={trackRef}
                onPointerDown={(e) => {
                  if ((e.target as HTMLElement).classList.contains("handle")) return;
                  setDrag("seek");
                  const rect = trackRef.current!.getBoundingClientRect();
                  const p = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
                  seek(p * duration);
                }}
              >
                <div
                  className="selection"
                  style={{ left: `${pct(start)}%`, width: `${pct(end) - pct(start)}%` }}
                />
                <div
                  className="handle handle-start"
                  style={{ left: `${pct(start)}%` }}
                  onPointerDown={(e) => {
                    e.stopPropagation();
                    setDrag("start");
                  }}
                />
                <div
                  className="handle handle-end"
                  style={{ left: `${pct(end)}%` }}
                  onPointerDown={(e) => {
                    e.stopPropagation();
                    setDrag("end");
                  }}
                />
                <div className="playhead" style={{ left: `${pct(current)}%` }} />
              </div>

              <div className="trim-times">
                <span>Start <strong>{fmt(start)}</strong></span>
                <button className="secondary small" onClick={togglePlay}>
                  {playing ? "Pause" : "Play selection"}
                </button>
                <span>End <strong>{fmt(end)}</strong></span>
              </div>
              <div className="trim-len">Selection: {fmt(end - start)} of {fmt(duration)}</div>
            </>
          )}

          <div className="controls">
            <RunBar job={job} label="Trim" onRun={run} disabled={!input || !canRun} self="trim" />
          </div>
        </div>
      )}
    </div>
  );
}
