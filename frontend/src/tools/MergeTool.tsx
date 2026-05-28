import { useEffect, useState } from "react";
import {
  Merge,
  SelectInputFiles,
  SelectOutputPath,
  SuggestOutput,
  Thumbnail,
} from "../../wailsjs/go/main/App";
import { RunBar } from "../components/RunBar";
import { useJob } from "../hooks/useJob";
import { ToolProps } from "../lib/nav";
import { baseName, extOf } from "../lib/util";
import { setDropHandler, clearDropHandler } from "../lib/dropTarget";

export function MergeTool({ initialInput }: ToolProps) {
  const [files, setFiles] = useState<string[]>(initialInput ? [initialInput] : []);
  const [thumbs, setThumbs] = useState<Record<string, string>>({});
  const job = useJob();

  // Fetch a thumbnail for any file we don't have one for yet.
  useEffect(() => {
    files.forEach((f) => {
      if (thumbs[f] === undefined) {
        setThumbs((t) => ({ ...t, [f]: "" })); // mark in-flight
        Thumbnail(f)
          .then((d) => setThumbs((t) => ({ ...t, [f]: d })))
          .catch(() => {});
      }
    });
  }, [files]);

  // Dropped files get appended to the list.
  useEffect(() => {
    const h = (paths: string[]) => setFiles((f) => [...f, ...paths]);
    setDropHandler(h);
    return () => clearDropHandler(h);
  }, []);

  async function add() {
    const picked = await SelectInputFiles();
    if (picked?.length) setFiles((f) => [...f, ...picked]);
  }

  const move = (i: number, d: number) =>
    setFiles((f) => {
      const j = i + d;
      if (j < 0 || j >= f.length) return f;
      const copy = [...f];
      [copy[i], copy[j]] = [copy[j], copy[i]];
      return copy;
    });

  const remove = (i: number) => setFiles((f) => f.filter((_, k) => k !== i));

  const run = () =>
    job.run(async () => {
      const out = await SelectOutputPath(await SuggestOutput(files[0], "_merged", extOf(files[0])));
      if (!out) return "";
      return Merge({ inputPaths: files, outputPath: out });
    });

  return (
    <div className="tool">
      <h2>Merge videos</h2>
      <p className="sub">
        Joins clips in order. Fast lossless concat — works best when all files share the
        same codec/resolution.
      </p>
      <div className="droplist">
        {files.length === 0 ? (
          <div className="placeholder" onClick={add} style={{ cursor: "pointer" }}>
            <strong>Add videos to merge</strong>
            <span>click to browse, or drop files here</span>
          </div>
        ) : (
          <ol className="filelist">
            {files.map((f, i) => (
              <li key={`${f}-${i}`}>
                <span className="num">{i + 1}</span>
                {thumbs[f] ? (
                  <img className="row-thumb" src={thumbs[f]} alt="" />
                ) : (
                  <span className="row-thumb thumb-loading" />
                )}
                <span className="name">{baseName(f)}</span>
                <span className="row-actions">
                  <button onClick={() => move(i, -1)} title="Move up">↑</button>
                  <button onClick={() => move(i, 1)} title="Move down">↓</button>
                  <button onClick={() => remove(i)} title="Remove">✕</button>
                </span>
              </li>
            ))}
          </ol>
        )}
      </div>
      <div className="controls">
        {files.length > 0 && (
          <button className="secondary" onClick={add}>
            Add more…
          </button>
        )}
        <RunBar job={job} label={`Merge ${files.length} videos`} onRun={run} disabled={files.length < 2} />
      </div>
    </div>
  );
}
