import { createContext, useContext } from "react";
import { ToolId } from "../tools/meta";

// Lets any component navigate to a tool, optionally pre-loading a video path.
// Used for Home drop-to-card and the "Next operation" chaining.
export type Nav = { openTool: (id: ToolId, input?: string) => void };

export const NavContext = createContext<Nav>({ openTool: () => {} });
export const useNav = () => useContext(NavContext);

// Shared prop for tool components.
export type ToolProps = { initialInput?: string };

// Tools that operate on a single video and can therefore be chained after any op.
export const CHAINABLE: ToolId[] = [
  "convert",
  "speed",
  "trim",
  "compress",
  "extract",
  "gif",
  "watermark",
  "rotate",
  "metadata",
];
