import { useState } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import DeployTabs from "./DeployTabs";
import QuickDeployTab from "./QuickDeployTab";
import ZipUploadTab from "./ZipUploadTab";
import GitHubRepoTab from "./GitHubRepoTab";
import { createGitHubDeployment, createZipDeployment, createPrebuiltDeployment } from "../../api/deployments";
import { useFriendCode } from "../../hooks/useFriendCode";
import { useToast } from "../shared/Toast";
import { extractNameFromGithubUrl } from "../../lib/utils";
import type { Deployment, DeployPreset } from "../../types/deployment";

interface DeployPanelProps {
  onDeployStarted: (deployment: Deployment) => void;
}

export default function DeployPanel({ onDeployStarted }: DeployPanelProps) {
  const [activeTab, setActiveTab] = useState("quick");
  const [isDeploying, setIsDeploying] = useState(false);
  const { friendCode } = useFriendCode();
  const { addToast } = useToast();

  const handlePresetDeploy = async (preset: DeployPreset, message?: string) => {
    setIsDeploying(true);
    try {
      const envVars: Record<string, string> = { ...preset.environmentVariables };
      if (message && preset.requiresTextInput) {
        envVars["VITE_CORVUS_MESSAGE"] = message;
      }

      let deployment: Deployment;

      if (preset.sourceType === "prebuilt" && preset.presetId) {
        // Prebuilt preset: skip clone + build, deploy from server-side prebuilt files
        deployment = await createPrebuiltDeployment({
          name: preset.requiresTextInput ? "Custom Message" : preset.name,
          presetId: preset.presetId,
          environmentVariables: Object.keys(envVars).length > 0 ? envVars : undefined,
          friendCode: friendCode || undefined,
        });
      } else if (preset.githubUrl) {
        // GitHub preset: clone + build from repo
        deployment = await createGitHubDeployment({
          name: preset.requiresTextInput ? "Custom Message" : preset.name,
          githubUrl: preset.githubUrl,
          branch: preset.branch,
          buildCommand: preset.buildCommand,
          outputDirectory: preset.outputDirectory,
          environmentVariables: Object.keys(envVars).length > 0 ? envVars : undefined,
          friendCode: friendCode || undefined,
        });
      } else {
        throw new Error("Preset is missing both presetId and githubUrl");
      }

      onDeployStarted(deployment);
    } catch (err) {
      addToast(err instanceof Error ? err.message : "Failed to create deployment", "error");
      setIsDeploying(false);
    }
  };

  const handleZipDeploy = async (file: File, outputDirectory: string, buildCommand: string) => {
    setIsDeploying(true);
    try {
      const deployment = await createZipDeployment({
        file, outputDirectory, buildCommand,
        friendCode: friendCode || undefined,
      });
      onDeployStarted(deployment);
    } catch (err) {
      addToast(err instanceof Error ? err.message : "Failed to create deployment", "error");
      setIsDeploying(false);
    }
  };

  const handleGitHubDeploy = async (repoUrl: string, branch: string, buildCommand: string, outputDirectory: string) => {
    setIsDeploying(true);
    try {
      const deployment = await createGitHubDeployment({
        name: extractNameFromGithubUrl(repoUrl),
        githubUrl: repoUrl, branch, buildCommand, outputDirectory,
        friendCode: friendCode || undefined,
      });
      onDeployStarted(deployment);
    } catch (err) {
      addToast(err instanceof Error ? err.message : "Failed to create deployment", "error");
      setIsDeploying(false);
    }
  };

  return (
    <div className="ink-card torn-edge-1" style={{ zIndex: 10, position: "relative", transform: "rotate(-0.4deg)" }}>
      <DeployTabs activeTab={activeTab} onTabChange={setActiveTab}>
        <Tabs.Content value="quick">
          <QuickDeployTab onDeploy={handlePresetDeploy} disabled={isDeploying} />
        </Tabs.Content>
        <Tabs.Content value="zip">
          <ZipUploadTab onDeploy={handleZipDeploy} disabled={isDeploying} />
        </Tabs.Content>
        <Tabs.Content value="github">
          <GitHubRepoTab onDeploy={handleGitHubDeploy} disabled={isDeploying} />
        </Tabs.Content>
      </DeployTabs>
    </div>
  );
}
