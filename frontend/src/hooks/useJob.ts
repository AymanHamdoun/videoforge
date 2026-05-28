import { useCallback, useState } from "react";
import { useJobEvents, JobState } from "./useJobEvents";

export type UseJob = {
  state: JobState;
  error: string;
  busy: boolean;
  run: (starter: () => Promise<string>) => Promise<void>;
};

/**
 * Drives one operation: calls a binding that returns a job ID, then tracks its
 * progress/result via Wails events. `error` captures synchronous failures
 * (validation, dialog cancel) before a job ID exists.
 */
export function useJob(): UseJob {
  const [jobId, setJobId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const state = useJobEvents(jobId);

  const run = useCallback(async (starter: () => Promise<string>) => {
    setError("");
    try {
      const id = await starter();
      if (id) setJobId(id);
    } catch (e) {
      setError(String(e));
    }
  }, []);

  const busy = state.status === "processing";
  return { state, error, busy, run };
}
