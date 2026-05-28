import { RevealInFolder } from "../../wailsjs/go/main/App";
import { UseJob } from "../hooks/useJob";

type Props = {
  job: UseJob;
  label: string;
  disabled?: boolean;
  onRun: () => void;
};

/** Shared run button + progress bar + result/error line. */
export function RunBar({ job, label, disabled, onRun }: Props) {
  const { state, error, busy } = job;
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

      {error && <p className="error">{error}</p>}
    </div>
  );
}
