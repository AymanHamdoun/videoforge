import { RevealInFolder } from "../../wailsjs/go/main/App";
import { UseJob } from "../hooks/useJob";
import { useNav, CHAINABLE } from "../lib/nav";
import { TOOL_META, ToolId } from "../tools/meta";

type Props = {
  job: UseJob;
  label: string;
  disabled?: boolean;
  onRun: () => void;
  self?: ToolId; // current tool, excluded from the next-op list
};

/** Shared run button + progress bar + result/error + "Next operation" chaining. */
export function RunBar({ job, label, disabled, onRun, self }: Props) {
  const { state, error, busy } = job;
  const { openTool } = useNav();

  const nextOps = TOOL_META.filter((t) => CHAINABLE.includes(t.id) && t.id !== self);

  return (
    <div className="runbar">
      <button className="primary" onClick={onRun} disabled={disabled || busy}>
        {busy ? "Working…" : label}
      </button>

      {state.status !== "idle" && (
        <div className="progress">
          <div className="bar">
            <div className="fill" style={{ width: `${state.percent}%` }} />
          </div>
          <div className="status">
            {state.status === "processing" && `${state.percent.toFixed(0)}%`}
            {state.status === "completed" && (
              <>
                ✅ Done —{" "}
                <button className="link" onClick={() => RevealInFolder(state.output!)}>
                  Show in folder
                </button>
              </>
            )}
            {state.status === "failed" && <span className="error">❌ {state.error}</span>}
            {state.status === "canceled" && "Canceled"}
          </div>
        </div>
      )}

      {state.status === "completed" && state.output && (
        <div className="nextops">
          <div className="nextops-title">Next operation — continue with this result:</div>
          <div className="nextops-list">
            {nextOps.map((t) => (
              <button key={t.id} className="nextop" onClick={() => openTool(t.id, state.output!)}>
                <span>{t.icon}</span>
                {t.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {error && <p className="error">{error}</p>}
    </div>
  );
}
