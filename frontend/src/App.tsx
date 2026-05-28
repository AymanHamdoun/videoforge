import { useEffect, useState } from "react";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";
import { fireDrop } from "./lib/dropTarget";
import { ConvertTool } from "./tools/ConvertTool";
import { SpeedTool } from "./tools/SpeedTool";
import { TrimTool } from "./tools/TrimTool";
import { CompressTool } from "./tools/CompressTool";
import { ExtractAudioTool } from "./tools/ExtractAudioTool";
import { MergeTool } from "./tools/MergeTool";
import { GifTool } from "./tools/GifTool";
import { WatermarkTool } from "./tools/WatermarkTool";
import { RotateTool } from "./tools/RotateTool";
import { MetadataTool } from "./tools/MetadataTool";
import "./App.css";

const TOOLS = [
  { id: "convert", label: "Convert", icon: "🔄", Comp: ConvertTool },
  { id: "speed", label: "Speed", icon: "⏩", Comp: SpeedTool },
  { id: "trim", label: "Trim", icon: "✂️", Comp: TrimTool },
  { id: "compress", label: "Compress", icon: "🗜️", Comp: CompressTool },
  { id: "extract", label: "Extract audio", icon: "🎵", Comp: ExtractAudioTool },
  { id: "merge", label: "Merge", icon: "🔗", Comp: MergeTool },
  { id: "gif", label: "GIF", icon: "🎞️", Comp: GifTool },
  { id: "watermark", label: "Watermark", icon: "💧", Comp: WatermarkTool },
  { id: "rotate", label: "Rotate / flip", icon: "🔁", Comp: RotateTool },
  { id: "metadata", label: "Metadata", icon: "ℹ️", Comp: MetadataTool },
] as const;

function App() {
  const [active, setActive] = useState<string>("convert");

  // One global file-drop listener routes the dropped paths to the active tool.
  useEffect(() => {
    OnFileDrop((_x, _y, paths) => fireDrop(paths), false);
    return () => OnFileDropOff();
  }, []);

  const Active = TOOLS.find((t) => t.id === active)!.Comp;

  return (
    <div className="app">
      <header className="topbar">
        <span className="logo">🎬 VideoForge</span>
      </header>
      <div className="body">
        <nav className="sidebar">
          {TOOLS.map((t) => (
            <button
              key={t.id}
              className={`navitem ${active === t.id ? "active" : ""}`}
              onClick={() => setActive(t.id)}
            >
              <span className="navicon">{t.icon}</span>
              {t.label}
            </button>
          ))}
        </nav>
        <main className="content">
          {/* Remount per tool so each starts with clean state. */}
          <Active key={active} />
        </main>
      </div>
    </div>
  );
}

export default App;
