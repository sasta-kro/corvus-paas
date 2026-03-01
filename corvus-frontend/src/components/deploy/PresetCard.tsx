import type { DeployPreset } from "../../types/deployment";

interface PresetCardProps {
  preset: DeployPreset;
  onDeploy: (preset: DeployPreset, message?: string) => void;
  disabled: boolean;
}

/** Individual preset card for Quick Deploy tab */
export default function PresetCard({ preset, onDeploy, disabled }: PresetCardProps) {
  return (
    <button
      onClick={() => {
        if (!disabled) {
          // For presets requiring text input, parent handles showing the modal
          onDeploy(preset);
        }
      }}
      disabled={disabled}
      className={`text-left border border-gray-200 rounded-lg p-5 transition-all cursor-pointer ${
        disabled
          ? "opacity-50 cursor-not-allowed"
          : "hover:border-black hover:shadow-sm"
      }`}
    >
      <div className="text-3xl mb-3">{preset.icon}</div>
      <h3 className="font-semibold text-base mb-1">{preset.name}</h3>
      <p className="text-sm text-gray-500">{preset.description}</p>
    </button>
  );
}

