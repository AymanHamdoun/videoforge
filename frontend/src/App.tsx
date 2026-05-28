import { useEffect, useState } from "react";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";
import { AppVersion } from "../wailsjs/go/main/App";
import { fireDrop } from "./lib/dropTarget";
import { TOOL_META, ToolId } from "./tools/meta";
import logo from "./assets/images/logo.png";
import { HomeTool } from "./tools/HomeTool";
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

const COMPONENTS: Record<Exclude<ToolId, "home">, () => JSX.Element> = {
  convert: ConvertTool,
  speed: SpeedTool,
  trim: TrimTool,
  compress: CompressTool,
  extract: ExtractAudioTool,
  merge: MergeTool,
  gif: GifTool,
  watermark: WatermarkTool,
  rotate: RotateTool,
  metadata: MetadataTool,
};

function App() {
  const [active, setActive] = useState<ToolId>("home");
  const [version, setVersion] = useState("");

  // One global file-drop listener routes the dropped paths to the active tool.
  useEffect(() => {
    OnFileDrop((_x, _y, paths) => fireDrop(paths), false);
    AppVersion().then(setVersion).catch(() => {});
    return () => OnFileDropOff();
  }, []);

  function renderActive() {
    if (active === "home") return <HomeTool onNavigate={setActive} />;
    const Active = COMPONENTS[active];
    return <Active key={active} />;
  }

  return (
    <div className="app">
      <header className="topbar">
        <img src={logo} className="logo-img" alt="" />
        <span className="logo">VideoForge</span>
      </header>
      <div className="body">
        <nav className="sidebar">
          <div className="navlist">
            {TOOL_META.map((t) => (
              <button
                key={t.id}
                className={`navitem ${active === t.id ? "active" : ""}`}
                onClick={() => setActive(t.id)}
              >
                <span className="navicon">{t.icon}</span>
                {t.label}
              </button>
            ))}
          </div>
          {version && <div className="version">v{version}</div>}
        </nav>
        <main className="content">
          {/* Remount per tool so each starts with clean state. */}
          {renderActive()}
        </main>
      </div>
    </div>
  );
}

export default App;
