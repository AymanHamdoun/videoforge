import { useState } from "react";
import { Activate, OpenPurchasePage } from "../../wailsjs/go/main/App";
import { license } from "../../wailsjs/go/models";
import logo from "../assets/images/logo.png";

export function ActivationGate({ onActivated }: { onActivated: (s: license.Status) => void }) {
  const [key, setKey] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function activate() {
    if (!key.trim()) return;
    setBusy(true);
    setError("");
    try {
      const status = await Activate(key.trim());
      onActivated(status);
    } catch (e) {
      setError(String(e).replace(/^Error:\s*/, ""));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="gate">
      <img src={logo} className="gate-logo" alt="VideoForge" />
      <h1>Activate VideoForge</h1>
      <p className="gate-sub">Enter your license key to unlock the app.</p>

      <textarea
        className="gate-input"
        placeholder="Paste your license key…"
        value={key}
        spellCheck={false}
        onChange={(e) => setKey(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) activate();
        }}
      />

      {error && <p className="gate-error">❌ {error}</p>}

      <button className="primary gate-activate" onClick={activate} disabled={busy || !key.trim()}>
        {busy ? "Checking…" : "Activate"}
      </button>

      <div className="gate-buy">
        Don't have a key?{" "}
        <button className="link" onClick={() => OpenPurchasePage()}>
          Buy a license
        </button>
      </div>
    </div>
  );
}
