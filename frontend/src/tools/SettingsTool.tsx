import { useEffect, useState } from "react";
import {
  Activate,
  Deactivate,
  GetPreferences,
  OpenPurchasePage,
  SavePreferences,
  SelectFolder,
} from "../../wailsjs/go/main/App";
import { main, settings } from "../../wailsjs/go/models";

type Props = {
  license: main.LicenseStatus;
  onLicenseChange: (s: main.LicenseStatus) => void;
};

export function SettingsTool({ license, onLicenseChange }: Props) {
  const [key, setKey] = useState("");
  const [licErr, setLicErr] = useState("");
  const [busy, setBusy] = useState(false);

  const [prefs, setPrefs] = useState<settings.Preferences | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    GetPreferences().then(setPrefs).catch(() => {});
  }, []);

  async function activate() {
    if (!key.trim()) return;
    setBusy(true);
    setLicErr("");
    try {
      const status = await Activate(key.trim());
      onLicenseChange(status);
      setKey("");
    } catch (e) {
      setLicErr(String(e).replace(/^Error:\s*/, ""));
    } finally {
      setBusy(false);
    }
  }

  async function deactivate() {
    await Deactivate();
    onLicenseChange({ activated: false } as main.LicenseStatus);
  }

  function expiryLabel() {
    if (license.perpetual) return "Perpetual license — never expires";
    const d = license.daysLeft;
    return `${d} day${d === 1 ? "" : "s"} left — expires ${license.expiry}`;
  }

  async function chooseFolder() {
    const dir = await SelectFolder();
    if (dir && prefs) setPrefs({ ...prefs, defaultOutputDir: dir });
  }

  async function savePrefs() {
    if (!prefs) return;
    const updated = await SavePreferences(prefs);
    setPrefs(updated);
    setSaved(true);
    setTimeout(() => setSaved(false), 1500);
  }

  return (
    <div className="tool settings">
      <h2>Settings</h2>
      <p className="sub">Manage your license and preferences.</p>

      {/* License */}
      <section className="card">
        <h3>License</h3>
        {license.activated ? (
          <>
            <div className="kv">
              <span>Licensed to</span>
              <strong>{license.name || "—"}</strong>
            </div>
            {license.email && (
              <div className="kv">
                <span>Email</span>
                <strong>{license.email}</strong>
              </div>
            )}
            <div className="kv">
              <span>Status</span>
              <strong className={!license.perpetual && license.daysLeft <= 14 ? "warn" : "ok"}>
                {expiryLabel()}
              </strong>
            </div>
            <button className="secondary danger" onClick={deactivate}>
              Deactivate this device
            </button>
          </>
        ) : (
          <p className="sub" style={{ textAlign: "left", margin: 0 }}>
            No active license.
          </p>
        )}

        <div className="relicense">
          <label className="field">
            <span>{license.activated ? "Replace license key" : "Enter license key"}</span>
            <textarea
              className="gate-input"
              style={{ maxWidth: "none" }}
              placeholder="Paste your license key…"
              value={key}
              spellCheck={false}
              onChange={(e) => setKey(e.target.value)}
            />
          </label>
          {licErr && <p className="error">❌ {licErr}</p>}
          <div className="row">
            <button className="primary" onClick={activate} disabled={busy || !key.trim()}>
              {busy ? "Checking…" : "Activate"}
            </button>
            <button className="link" onClick={() => OpenPurchasePage()}>
              Buy a license
            </button>
          </div>
        </div>
      </section>

      {/* Preferences */}
      {prefs && (
        <section className="card">
          <h3>Preferences</h3>
          <div className="field">
            <span>Default output folder</span>
            <div className="row">
              <input readOnly value={prefs.defaultOutputDir || "Ask each time"} />
              <button className="secondary" onClick={chooseFolder}>
                Choose…
              </button>
              {prefs.defaultOutputDir && (
                <button className="link" onClick={() => setPrefs({ ...prefs, defaultOutputDir: "" })}>
                  Clear
                </button>
              )}
            </div>
          </div>

          <label className="field">
            <span>Max concurrent jobs (applies after restart)</span>
            <input
              type="number"
              min={1}
              max={8}
              value={prefs.maxConcurrentJobs}
              onChange={(e) =>
                setPrefs({ ...prefs, maxConcurrentJobs: Math.max(1, Math.min(8, +e.target.value || 1)) })
              }
            />
          </label>

          <div className="row">
            <button className="primary" onClick={savePrefs}>
              Save preferences
            </button>
            {saved && <span className="ok">Saved ✓</span>}
          </div>
        </section>
      )}
    </div>
  );
}
