// Single source of truth for the tool list. Pure data (no component imports) so
// both the sidebar (App.tsx) and the Home grid can use it without import cycles.

import homeIcon from "../assets/images/logo.png";
import convertIcon from "../assets/icons/convert.png";
import speedIcon from "../assets/icons/speed.png";
import trimIcon from "../assets/icons/trim.png";
import compressIcon from "../assets/icons/compress.png";
import extractIcon from "../assets/icons/extract.png";
import mergeIcon from "../assets/icons/merge.png";
import gifIcon from "../assets/icons/gif.png";
import watermarkIcon from "../assets/icons/watermark.png";
import rotateIcon from "../assets/icons/rotate.png";
import metadataIcon from "../assets/icons/metadata.png";
import settingsIcon from "../assets/icons/settings.png";

export type ToolId =
  | "home"
  | "convert"
  | "speed"
  | "trim"
  | "compress"
  | "extract"
  | "merge"
  | "gif"
  | "watermark"
  | "rotate"
  | "metadata"
  | "settings";

export type ToolMeta = {
  id: ToolId;
  label: string;
  // Emoji glyph fallback (used when no image is provided, e.g. Home).
  icon: string;
  // Imported icon image. When set, UI prefers this over `icon`.
  iconImg?: string;
  desc: string; // short, plain-English description for the Home grid
};

export const TOOL_META: ToolMeta[] = [
  { id: "home", label: "Home", icon: "🏠", iconImg: homeIcon, desc: "Overview and quick actions" },
  { id: "convert", label: "Convert", icon: "🔄", iconImg: convertIcon, desc: "Change the file format, e.g. MOV → MP4" },
  { id: "speed", label: "Speed", icon: "⏩", iconImg: speedIcon, desc: "Speed up or slow down a video" },
  { id: "trim", label: "Trim", icon: "✂️", iconImg: trimIcon, desc: "Cut out a section (instant, no quality loss)" },
  { id: "compress", label: "Compress", icon: "🗜️", iconImg: compressIcon, desc: "Make the file smaller" },
  { id: "extract", label: "Extract audio", icon: "🎵", iconImg: extractIcon, desc: "Save just the soundtrack" },
  { id: "merge", label: "Merge", icon: "🔗", iconImg: mergeIcon, desc: "Join several clips into one" },
  { id: "gif", label: "GIF", icon: "🎞️", iconImg: gifIcon, desc: "Turn a clip into an animated GIF" },
  { id: "watermark", label: "Watermark", icon: "💧", iconImg: watermarkIcon, desc: "Overlay a logo or image" },
  { id: "rotate", label: "Rotate / flip", icon: "🔁", iconImg: rotateIcon, desc: "Reorient the video" },
  { id: "metadata", label: "Metadata", icon: "ℹ️", iconImg: metadataIcon, desc: "Inspect format and stream details" },
  { id: "settings", label: "Settings", icon: "⚙️", iconImg: settingsIcon, desc: "License and preferences" },
];

// Glossary shown on the Home page — explains the jargon used across the tools.
export type Term = { term: string; def: string };

export const GLOSSARY: Term[] = [
  {
    term: "Container",
    def: "The file type / wrapper — .mp4, .mkv, .mov, .webm. It decides where the video plays, not how good it looks. Converting between containers is what the Convert tool does.",
  },
  {
    term: "Codec",
    def: "How the video or audio is compressed inside the container (e.g. H.264 / H.265 for video, AAC / MP3 for audio). One container holds one or more codec 'streams'.",
  },
  {
    term: "CRF (Constant Rate Factor)",
    def: "A quality dial for H.264. Lower = better quality and a bigger file. 18 ≈ visually lossless, 23 is the default, 28+ gets noticeably smaller with visible quality loss.",
  },
  {
    term: "Preset",
    def: "How hard the encoder works. Slower presets (slow, veryslow) squeeze the file smaller at the same quality but take longer; faster presets finish quickly but produce larger files.",
  },
  {
    term: "Bitrate",
    def: "How much data is used per second of video. Higher bitrate = more detail and larger files.",
  },
  {
    term: "Resolution",
    def: "The frame size in pixels, width × height. 1920×1080 is '1080p', 1280×720 is '720p'. Bigger isn't always better — it just means more pixels.",
  },
  {
    term: "FPS (frames per second)",
    def: "How many images are shown each second. 24–30 is normal for video; 60 looks extra smooth. GIFs usually use 10–15 to stay small.",
  },
  {
    term: "Stream copy vs re-encode",
    def: "Trim and Merge copy the existing data as-is — instant and lossless. Most other tools re-encode the video, which is slower and slightly changes quality, but is required to actually transform it.",
  },
];
