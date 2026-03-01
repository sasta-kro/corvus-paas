import { useState, useCallback } from "react";
import PresetCard from "./PresetCard";
import MessageInputModal from "./MessageInputModal";
import { DEPLOY_PRESETS } from "../../config/constants";
import type { DeployPreset } from "../../types/deployment";

interface QuickDeployTabProps {
  onDeploy: (preset: DeployPreset, message?: string) => void;
  disabled: boolean;
}

/** Quick Deploy tab — renders preset cards in a grid */
export default function QuickDeployTab({ onDeploy, disabled }: QuickDeployTabProps) {
  const [messageModalOpen, setMessageModalOpen] = useState(false);
  const [selectedPreset, setSelectedPreset] = useState<DeployPreset | null>(null);

  const handlePresetClick = useCallback(
    (preset: DeployPreset) => {
      if (preset.requiresTextInput) {
        setSelectedPreset(preset);
        setMessageModalOpen(true);
      } else {
        onDeploy(preset);
      }
    },
    [onDeploy]
  );

  const handleMessageSubmit = useCallback(
    (message: string) => {
      if (selectedPreset) {
        onDeploy(selectedPreset, message);
        setMessageModalOpen(false);
        setSelectedPreset(null);
      }
    },
    [selectedPreset, onDeploy]
  );

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {DEPLOY_PRESETS.map((preset) => (
          <PresetCard
            key={preset.id}
            preset={preset}
            onDeploy={handlePresetClick}
            disabled={disabled}
          />
        ))}
      </div>

      <MessageInputModal
        isOpen={messageModalOpen}
        onClose={() => {
          setMessageModalOpen(false);
          setSelectedPreset(null);
        }}
        onSubmit={handleMessageSubmit}
      />
    </>
  );
}

