import { TOOL_META, GLOSSARY, ToolId } from "./meta";

export function HomeTool({ onNavigate }: { onNavigate: (id: ToolId) => void }) {
  const actions = TOOL_META.filter((t) => t.id !== "home");

  return (
    <div className="home">
      <section className="hero">
        <div className="hero-logo">🎬</div>
        <h1>VideoForge</h1>
        <p>A fast, offline video toolkit. Everything runs on your machine — nothing is uploaded.</p>
      </section>

      <section>
        <h2 className="section-title">Quick actions</h2>
        <div className="action-grid">
          {actions.map((t) => (
            <button key={t.id} className="action-card" onClick={() => onNavigate(t.id)}>
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
