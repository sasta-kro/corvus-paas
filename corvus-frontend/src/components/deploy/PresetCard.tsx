import type { DeployPreset } from "../../types/deployment";

function PresetIcon({ name }: { name: string }) {
  const s = { display: "block" };
  switch (name) {
    case "vite":
      return (
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" style={s}>
          {/* Ink lightning bolt */}
          <path d="M13 2L4.5 13.5H11L10 22L19.5 10.5H13L13 2Z" stroke="var(--sumi)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" fill="none" />
        </svg>
      );
    case "react":
      return (
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" style={s}>
          {/* Ink atom — three orbital ellipses + center dot */}
          <ellipse cx="12" cy="12" rx="10" ry="4" stroke="var(--sumi)" strokeWidth="1.5" fill="none" />
          <ellipse cx="12" cy="12" rx="10" ry="4" stroke="var(--sumi)" strokeWidth="1.5" fill="none" transform="rotate(60 12 12)" />
          <ellipse cx="12" cy="12" rx="10" ry="4" stroke="var(--sumi)" strokeWidth="1.5" fill="none" transform="rotate(120 12 12)" />
          <circle cx="12" cy="12" r="2" fill="var(--sumi)" />
        </svg>
      );
    case "crow":
      return (
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" style={s}>
          {/* Ink crow silhouette — perched bird */}
          <path d="M4 18C4 18 5 14 7 12C9 10 10 9 10 7C10 5 9 3 9 3C9 3 12 5 13 7C14 9 14 10 15 11C16 12 18 13 20 13C20 13 18 15 15 15C14 15 13 15.5 12 16.5C11 17.5 10 19 10 19L8 19C8 19 8 17 7 16C6 15 4 18 4 18Z" stroke="var(--sumi)" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" fill="none" />
          <circle cx="10.5" cy="6.5" r="0.8" fill="var(--sumi)" />
        </svg>
      );
    case "brush":
      return (
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" style={s}>
          {/* Ink calligraphy brush */}
          <path d="M18 3C18 3 14 7 12 11C11 13 10.5 14 10 15C9.5 16 9 17 9 18C9 19.5 10 21 12 21C14 21 15 19.5 15 18C15 17 14.5 16 14 15.5" stroke="var(--sumi)" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" fill="none" />
          <path d="M12 21C11 21 9.5 20.5 9 18" stroke="var(--sumi)" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      );
    default:
      return (
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" style={s}>
          <circle cx="12" cy="12" r="4" stroke="var(--sumi)" strokeWidth="1.8" fill="none" />
        </svg>
      );
  }
}

interface PresetCardProps {
  preset: DeployPreset;
  onDeploy: (preset: DeployPreset, message?: string) => void;
  disabled: boolean;
}

export default function PresetCard({ preset, onDeploy, disabled }: PresetCardProps) {
  return (
    <button
      onClick={() => !disabled && onDeploy(preset)}
      disabled={disabled}
      className="preset-card torn-edge-sm"
    >
      <div className="mb-3"><PresetIcon name={preset.icon} /></div>
      <h3
        className="font-brush text-base mb-1"
        style={{ color: "var(--sumi)" }}
      >
        {preset.name}
      </h3>
      <p style={{ color: "var(--sumi-light)", fontSize: "0.85rem" }}>
        {preset.description}
      </p>
    </button>
  );
}
