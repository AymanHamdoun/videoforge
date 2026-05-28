import { useState } from "react";
import { VideoInput } from "../components/VideoInput";
import { ffmpeg } from "../../wailsjs/go/models";
import { ToolProps } from "../lib/nav";

function fmtBytes(s: string): string {
  const n = parseInt(s || "0", 10);
  if (!n) return "—";
  const u = ["B", "KB", "MB", "GB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(1)} ${u[i]}`;
}

export function MetadataTool({ initialInput }: ToolProps) {
  const [meta, setMeta] = useState<ffmpeg.MediaInfo | null>(null);
  const [path, setPath] = useState(initialInput ?? "");

  const dur = meta ? parseFloat(meta.format.duration || "0") : 0;

  return (
    <div className="tool">
      <h2>Metadata</h2>
      <p className="sub">Inspect format and stream details (read-only).</p>
      <VideoInput
        path={path}
        onChange={(p, m) => {
          setPath(p);
          setMeta(m);
        }}
      />
      {meta && (
        <div className="metadata">
          <table>
            <tbody>
              <tr><td>Format</td><td>{meta.format.format_name}</td></tr>
              <tr><td>Duration</td><td>{dur.toFixed(2)} s</td></tr>
              <tr><td>Size</td><td>{fmtBytes(meta.format.size)}</td></tr>
              <tr><td>Bitrate</td><td>{meta.format.bit_rate ? `${Math.round(+meta.format.bit_rate / 1000)} kbps` : "—"}</td></tr>
            </tbody>
          </table>
          <h3>Streams</h3>
          <table>
            <thead>
              <tr><th>#</th><th>Type</th><th>Codec</th><th>Details</th></tr>
            </thead>
            <tbody>
              {meta.streams.map((s) => (
                <tr key={s.index}>
                  <td>{s.index}</td>
                  <td>{s.codec_type}</td>
                  <td>{s.codec_name}</td>
                  <td>{s.codec_type === "video" && s.width ? `${s.width}×${s.height} @ ${s.r_frame_rate}` : ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
