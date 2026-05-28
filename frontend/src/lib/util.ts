/** Returns the file extension including the dot, defaulting to .mp4. */
export function extOf(path: string): string {
  const i = path.lastIndexOf(".");
  return i >= 0 ? path.slice(i) : ".mp4";
}

export const baseName = (p: string) => p.split(/[\\/]/).pop() || p;
