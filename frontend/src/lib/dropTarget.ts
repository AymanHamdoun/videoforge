// Tiny registry so the single global Wails OnFileDrop handler (set up once in
// App) can route dropped file paths to whichever tool is currently mounted.
type Handler = (paths: string[]) => void;

let current: Handler | null = null;

export function setDropHandler(h: Handler) {
  current = h;
}

export function clearDropHandler(h: Handler) {
  if (current === h) current = null;
}

export function fireDrop(paths: string[]) {
  current?.(paths);
}
