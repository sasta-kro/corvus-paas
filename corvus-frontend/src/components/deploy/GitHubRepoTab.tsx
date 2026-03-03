import { useState } from "react";
import DeployButton from "./DeployButton";
import BuildImageWarningModal from "./BuildImageWarningModal";
import { isValidGithubUrl } from "../../lib/utils";

/** Returns true if the command looks like it only uses the Node.js toolchain. */
function isNodeCommand(cmd: string): boolean {
  if (!cmd.trim()) return true; // empty = no build step, always fine
  // Split on shell separators so we check every segment individually
  const segments = cmd.split(/&&|;|\|/).map((s) => s.trim()).filter(Boolean);
  const nodeTools = /^(npm|yarn|pnpm|npx|node|corepack)\b/;
  return segments.every((seg) => nodeTools.test(seg));
}

interface GitHubRepoTabProps {
  onDeploy: (repoUrl: string, branch: string, buildCommand: string, outputDirectory: string) => void;
  disabled: boolean;
}

export default function GitHubRepoTab({ onDeploy, disabled }: GitHubRepoTabProps) {
  const [repoUrl, setRepoUrl] = useState("");
  const [branch, setBranch] = useState("main");
  const [buildCommand, setBuildCommand] = useState("");
  const [outputDirectory, setOutputDirectory] = useState("dist");
  const [urlError, setUrlError] = useState("");
  const [buildError, setBuildError] = useState("");
  const [showBuildWarning, setShowBuildWarning] = useState(false);

  const validate = (): boolean => {
    let valid = true;
    if (!isValidGithubUrl(repoUrl)) { setUrlError("Only public GitHub repositories are supported."); valid = false; } else { setUrlError(""); }
    setBuildError("");
    return valid;
  };

  const doDeploy = () => onDeploy(repoUrl.trim(), branch.trim() || "main", buildCommand.trim(), outputDirectory.trim() || ".");

  const handleDeploy = () => {
    if (!validate()) return;
    if (!isNodeCommand(buildCommand)) {
      setShowBuildWarning(true);
      return;
    }
    doDeploy();
  };

  return (
    <>
      <BuildImageWarningModal
        isOpen={showBuildWarning}
        onClose={() => setShowBuildWarning(false)}
        onDeployAnyway={doDeploy}
      />
      <div className="space-y-5">
        <div>
          <label htmlFor="gh-repo-url" className="ink-label">Repository URL</label>
          <input id="gh-repo-url" type="url" value={repoUrl}
            onChange={(e) => { setRepoUrl(e.target.value); if (urlError) setUrlError(""); }}
            placeholder="https://github.com/user/repo" disabled={disabled} className="ink-input" />
          {urlError && <p style={{ color: "var(--vermillion)", fontSize: "0.8rem", marginTop: "0.3rem" }}>{urlError}</p>}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-5">
          <div>
            <label htmlFor="gh-branch" className="ink-label">Branch</label>
            <input id="gh-branch" type="text" value={branch} onChange={(e) => setBranch(e.target.value)}
              placeholder="main" disabled={disabled} className="ink-input" />
          </div>
          <div>
            <label htmlFor="gh-build-cmd" className="ink-label">Build Command</label>
            <input id="gh-build-cmd" type="text" value={buildCommand}
              onChange={(e) => { setBuildCommand(e.target.value); if (buildError) setBuildError(""); }}
              placeholder="npm ci && npm run build" disabled={disabled} className="ink-input" style={{ fontFamily: "monospace", fontSize: "0.75rem" }} />
            {buildError && <p style={{ color: "var(--vermillion)", fontSize: "0.8rem", marginTop: "0.3rem" }}>{buildError}</p>}
          </div>
          <div>
            <label htmlFor="gh-output-dir" className="ink-label">Output Directory</label>
            <input id="gh-output-dir" type="text" value={outputDirectory} onChange={(e) => setOutputDirectory(e.target.value)}
              placeholder="e.g., dist, build, ." disabled={disabled} className="ink-input" />
          </div>
        </div>

        <DeployButton onClick={handleDeploy} disabled={disabled || !repoUrl.trim()} loading={disabled} />
      </div>
    </>
  );
}
