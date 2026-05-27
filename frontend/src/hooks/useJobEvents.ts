import { useEffect, useState } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { jobs } from "../../wailsjs/go/models";

type ProgressEvent = { id: string; percent: number };

export type JobState = {
  status: "idle" | "processing" | "completed" | "failed" | "canceled";
  percent: number;
  output?: string;
  error?: string;
};

const initial: JobState = { status: "idle", percent: 0 };

/**
 * Subscribes to the Go core's "job:progress" / "job:done" events for one job ID.
 * Replaces the old SSE EventSource. Listeners are cleaned up on unmount / ID change.
 */
export function useJobEvents(jobId: string | null): JobState {
  const [state, setState] = useState<JobState>(initial);

  useEffect(() => {
    if (!jobId) {
      setState(initial);
      return;
    }
    setState({ status: "processing", percent: 0 });

    const offProgress = EventsOn("job:progress", (p: ProgressEvent) => {
      if (p.id === jobId) setState((s) => ({ ...s, status: "processing", percent: p.percent }));
    });

    const offDone = EventsOn("job:done", (job: jobs.Job) => {
      if (job.id !== jobId) return;
      setState({
        status: job.status as JobState["status"],
        percent: job.progress,
        output: job.outputFile,
        error: job.error,
      });
    });

    return () => {
      offProgress();
      offDone();
    };
  }, [jobId]);

  return state;
}
