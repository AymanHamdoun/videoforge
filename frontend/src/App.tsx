import { useEffect, useState } from "react";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";
import { AppVersion, LicenseStatus } from "../wailsjs/go/main/App";
import { license } from "../wailsjs/go/models";
import { fireDrop } from "./lib/dropTarget";
import { NavContext, ToolProps } from "./lib/nav";
import { TOOL_META, ToolId } from "./tools/meta";
import logo from "./assets/images/logo.png";
import { ActivationGate } from "./components/ActivationGate";
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
import { SettingsTool } from "./tools/SettingsTool";
import "./App.css";

const COMPONENTS: Record<Exclude<ToolId, "home" | "settings">, (p: ToolProps) => JSX.Element> = {
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
  const [pendingInput, setPendingInput] = useState<string | undefined>(undefined);
  // navSeq forces a remount when re-opening the same tool with a new input.
  const [navSeq, setNavSeq] = useState(0);
  const [version, setVersion] = useState("");
  const [license, setLicense] = useState<license.Status | null>(null);

  function openTool(id: ToolId, input?: string) {
    setPendingInput(input);
    setActive(id);
    setNavSeq((n) => n + 1);
  }

  useEffect(() => {
    // Drop on a Home card → open that tool with the video; otherwise route to
    // the active tool's input zone.
    OnFileDrop((x, y, paths) => {
      if (!paths.length) return;
      const el = document.elementFromPoint(x, y)?.closest("[data-drop-tool]");
      if (el) {
        openTool(el.getAttribute("data-drop-tool") as ToolId, paths[0]);
        return;
      }
      fireDrop(paths);
    }, false);
    AppVersion().then(setVersion).catch(() => {});
    LicenseStatus()
      .then(setLicense)
      .catch(() => setLicense({ activated: false } as license.Status));
    return () => OnFileDropOff();
  }, []);

  if (license === null) return <div className="app" />;
  if (!license.activated) return <ActivationGate onActivated={setLicense} />;

  function renderActive() {
    if (active === "home") return <HomeTool />;
    if (active === "settings") return <SettingsTool license={license!} onLicenseChange={setLicense} />;
    const Active = COMPONENTS[active];
    return <Active key={`${active}-${navSeq}`} initialInput={pendingInput} />;
  }

  return (
    <NavContext.Provider value={{ openTool }}>
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
                  onClick={() => openTool(t.id)}
                >
                  <span className={`navicon ${t.id === "home" ? "navicon-tight" : ""}`}>
                    {t.iconImg ? <img src={t.iconImg} alt="" /> : t.icon}
                  </span>
                  {t.label}
                </button>
              ))}
            </div>
            <div className="sidebar-footer">
              {license.name && <div className="licensee">Licensed to {license.name}</div>}
              {version && <div className="version">v{version}</div>}
            </div>
          </nav>
          <main className="content">{renderActive()}</main>
        </div>
      </div>
    </NavContext.Provider>
  );
}

export default App;
