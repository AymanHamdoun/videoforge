import { useEffect, useState } from "react";
import { SelectInputFile, GetMetadata } from "../../wailsjs/go/main/App";
import { ffmpeg } from "../../wailsjs/go/models";
import { setDropHandler, clearDropHandler } from "../lib/dropTarget";

type Props = {
  path: string;
  onChange: (path: string, meta: ffmpeg.MediaInfo | null) => void;
};

const baseName = (p: string) => p.split(/[\\/]/).pop() || p;

/** Single-video picker: click to browse or drag-and-drop, shows metadata. */
export function VideoInput({ path, onChange }: Props) {
  const [meta, setMeta] = useState<ffmpeg.MediaInfo | null>(null);

  async function load(p: string) {
    try {
      const m = await GetMetadata(p);
      setMeta(m);
      onChange(p, m);
    } catch {
      setMeta(null);
      onChange(p, null);
    }
  }

  useEffect(() => {
    const h = (paths: string[]) => paths[0] && load(paths[0]);
    setDropHandler(h);
    return () => clearDropHandler(h);
  }, []);

  async function browse() {
    const p = await SelectInputFile();
    if (p) load(p);
  }

  const v = meta?.streams.find((s) => s.codec_type === "video");
  const dur = meta ? parseFloat(meta.format.duration || "0") : 0;

  return (
    <div className={`drop ${path ? "has-file" : ""}`} onClick={browse}>
      {path ? (
        <div className="fileinfo">
          <div className="filename">{baseName(path)}</div>
          {v && (
            <div className="meta">
              {v.codec_name} · {v.width}×{v.height} · {dur.toFixed(1)}s
            </div>
          )}
        </div>
      ) : (
        <div className="placeholder">
          <strong>Drop a video here</strong>
          <span>or click to browse</span>
        </div>
      )}
    </div>
  );
}
