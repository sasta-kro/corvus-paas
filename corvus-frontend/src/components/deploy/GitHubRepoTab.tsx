import { useState } from "react";
import DeployButton from "./DeployButton";
import { isValidGithubUrl } from "../../lib/utils";

interface GitHubRepoTabProps {
  onDeploy: (
    repoUrl: string,
    branch: string,
    buildCommand: string,
    outputDirectory: string
  ) => void;
  disabled: boolean;
}

/** GitHub Repo tab — URL, branch, build command, output directory fields */
export default function GitHubRepoTab({ onDeploy, disabled }: GitHubRepoTabProps) {
  const [repoUrl, setRepoUrl] = useState("");
  const [branch, setBranch] = useState("main");
  const [buildCommand, setBuildCommand] = useState("");
  const [outputDirectory, setOutputDirectory] = useState("dist");
  const [urlError, setUrlError] = useState("");
  const [buildError, setBuildError] = useState("");

  const validate = (): boolean => {
    let valid = true;

    if (!isValidGithubUrl(repoUrl)) {
      setUrlError("Only public GitHub repositories are supported.");
      valid = false;
    } else {
      setUrlError("");
    }

    if (!buildCommand.trim()) {
      setBuildError("Build command is required for GitHub deployments.");
      valid = false;
    } else {
      setBuildError("");
    }

    return valid;
  };

  const handleDeploy = () => {
    if (validate()) {
      onDeploy(repoUrl.trim(), branch.trim() || "main", buildCommand.trim(), outputDirectory.trim() || "dist");
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <label htmlFor="gh-repo-url" className="block text-sm font-medium mb-1">
          Repository URL
        </label>
        <input
          id="gh-repo-url"
          type="url"
          value={repoUrl}
          onChange={(e) => {
            setRepoUrl(e.target.value);
            if (urlError) setUrlError("");
          }}
          placeholder="https://github.com/user/repo"
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
        />
        {urlError && <p className="text-red-500 text-xs mt-1">{urlError}</p>}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div>
          <label htmlFor="gh-branch" className="block text-sm font-medium mb-1">
            Branch
          </label>
          <input
            id="gh-branch"
            type="text"
            value={branch}
            onChange={(e) => setBranch(e.target.value)}
            placeholder="main"
            disabled={disabled}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
          />
        </div>
        <div>
          <label htmlFor="gh-build-cmd" className="block text-sm font-medium mb-1">
            Build Command
          </label>
          <input
            id="gh-build-cmd"
            type="text"
            value={buildCommand}
            onChange={(e) => {
              setBuildCommand(e.target.value);
              if (buildError) setBuildError("");
            }}
            placeholder="npm ci && npm run build"
            disabled={disabled}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
          />
          {buildError && (
            <p className="text-red-500 text-xs mt-1">{buildError}</p>
          )}
        </div>
        <div>
          <label htmlFor="gh-output-dir" className="block text-sm font-medium mb-1">
            Output Directory
          </label>
          <input
            id="gh-output-dir"
            type="text"
            value={outputDirectory}
            onChange={(e) => setOutputDirectory(e.target.value)}
            placeholder="dist"
            disabled={disabled}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
          />
        </div>
      </div>

      <DeployButton
        onClick={handleDeploy}
        disabled={disabled || !repoUrl.trim() || !buildCommand.trim()}
        loading={disabled}
      />
    </div>
  );
}

