import { TOOL_META, GLOSSARY } from "./meta";
import { useNav } from "../lib/nav";
import logo from "../assets/images/logo.png";

export function HomeTool() {
  const { openTool } = useNav();
  const actions = TOOL_META.filter((t) => t.id !== "home");

  return (
    <div className="home">
      <section className="hero">
        <img src={logo} className="hero-logo-img" alt="VideoForge" />
        <h1>VideoForge</h1>
        <p>A fast, offline video toolkit. Everything runs on your machine — nothing is uploaded.</p>
      </section>

      <section>
        <h2 className="section-title">Quick actions</h2>
        <p className="drop-tip">
          <span className="drop-tip-icon">⤵</span> Drag a video onto any tool to start with it
          loaded — or just click to open.
        </p>
        <div className="action-grid">
          {actions.map((t) => (
            <button
              key={t.id}
              className="action-card"
              data-drop-tool={t.id}
              onClick={() => openTool(t.id)}
              onDragEnter={(e) => e.currentTarget.classList.add("dragover")}
              onDragOver={(e) => e.preventDefault()}
              onDragLeave={(e) => {
                if (!e.currentTarget.contains(e.relatedTarget as Node)) {
                  e.currentTarget.classList.remove("dragover");
                }
              }}
              onDrop={(e) => e.currentTarget.classList.remove("dragover")}
            >
              <span className="action-drop-hint">Drop here</span>
              <span className="action-icon">{t.icon}</span>
              <span className="action-label">{t.label}</span>
              <span className="action-desc">{t.desc}</span>
            </button>
          ))}
        </div>
      </section>

      <section>
        <h2 className="section-title">Legend — what the terms mean</h2>
        <p className="legend-intro">
          New to video settings? Here's the jargon you'll see around the app, in plain English.
        </p>
        <dl className="glossary">
          {GLOSSARY.map((g) => (
            <div className="term" key={g.term}>
              <dt>{g.term}</dt>
              <dd>{g.def}</dd>
            </div>
          ))}
        </dl>
      </section>
    </div>
  );
}
