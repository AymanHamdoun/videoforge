import { useEffect, useRef, useState } from "react";
import { SelectInputFile, GetMetadata, Thumbnail } from "../../wailsjs/go/main/App";
import { ffmpeg } from "../../wailsjs/go/models";
import { setDropHandler, clearDropHandler } from "../lib/dropTarget";
import { baseName } from "../lib/util";

type Props = {
  path: string;
  onChange: (path: string, meta: ffmpeg.MediaInfo | null) => void;
};

/** Single-video picker: click/drag to choose, shows a thumbnail + metadata.
 *  Loads whenever `path` changes (so chained/preset inputs load too). */
export function VideoInput({ path, onChange }: Props) {
  const [meta, setMeta] = useState<ffmpeg.MediaInfo | null>(null);
  const [thumb, setThumb] = useState("");
  const loaded = useRef("");

  useEffect(() => {
    if (!path || path === loaded.current) return;
    loaded.current = path;
    setMeta(null);
    setThumb("");
    GetMetadata(path)
      .then((m) => {
        setMeta(m);
        onChange(path, m);
      })
      .catch(() => onChange(path, null));
    Thumbnail(path).then(setThumb).catch(() => {});
  }, [path]);

  // Drag-and-drop into this zone (routed from the global handler).
  useEffect(() => {
    const h = (paths: string[]) => paths[0] && onChange(paths[0], null);
    setDropHandler(h);
    return () => clearDropHandler(h);
  }, [onChange]);

  async function browse() {
    const p = await SelectInputFile();
    if (p) onChange(p, null);
  }

  const v = meta?.streams.find((s) => s.codec_type === "video");
  const dur = meta ? parseFloat(meta.format.duration || "0") : 0;

  return (
    <div className={`drop ${path ? "has-file" : ""}`} onClick={browse}>
      {path ? (
        <div className="fileinfo">
          <div className="thumb-wrap">
            {thumb ? <img className="thumb" src={thumb} alt="" /> : <div className="thumb thumb-loading" />}
          </div>
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
