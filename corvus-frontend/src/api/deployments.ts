/** Deployment API functions — maps to backend endpoints */
import type { Deployment } from "../types/deployment";
import { apiGet, apiDelete, apiPostFormData, apiPost } from "./client";
import {
  extractNameFromFilename,
  normalizeGithubUrl,
} from "../lib/utils";

/** Creates a new deployment from a zip file upload */
export async function createZipDeployment(params: {
  file: File;
  outputDirectory: string;
  buildCommand: string;
  friendCode?: string;
}): Promise<Deployment> {
  const formData = new FormData();
  formData.append("name", extractNameFromFilename(params.file.name));
  formData.append("source_type", "zip");
  formData.append("file", params.file);
  formData.append("output_directory", params.outputDirectory || ".");
  formData.append("build_command", params.buildCommand || "");
  if (params.friendCode) {
    formData.append("friend_code", params.friendCode);
  }
  return apiPostFormData<Deployment>("/api/deployments", formData);
}

/** Creates a new deployment from a GitHub repo URL */
export async function createGitHubDeployment(params: {
  name: string;
  githubUrl: string;
  branch: string;
  buildCommand: string;
  outputDirectory: string;
  environmentVariables?: Record<string, string>;
  friendCode?: string;
}): Promise<Deployment> {
  const formData = new FormData();
  formData.append("name", params.name);
  formData.append("source_type", "github");
  formData.append("github_url", normalizeGithubUrl(params.githubUrl));
  formData.append("branch", params.branch || "main");
  formData.append("build_command", params.buildCommand);
  formData.append("output_directory", params.outputDirectory || "dist");
  if (params.environmentVariables) {
    formData.append(
      "environment_variables",
      JSON.stringify(params.environmentVariables)
    );
  }
  if (params.friendCode) {
    formData.append("friend_code", params.friendCode);
  }
  return apiPostFormData<Deployment>("/api/deployments", formData);
}

/** Fetches a single deployment by ID */
export async function getDeployment(id: string): Promise<Deployment> {
  return apiGet<Deployment>(`/api/deployments/${id}`);
}

/** Deletes a deployment by ID */
export async function deleteDeployment(id: string): Promise<void> {
  return apiDelete<void>(`/api/deployments/${id}`);
}

/** Triggers a redeploy for an existing deployment */
export async function redeployDeployment(id: string): Promise<Deployment> {
  return apiPost<Deployment>(`/api/deployments/${id}/redeploy`);
}

/** Validates a friend code */
export async function validateFriendCode(
  code: string
): Promise<{ valid: boolean }> {
  return apiGet<{ valid: boolean }>(`/api/validate-code?code=${encodeURIComponent(code)}`);
}

/** Creates a deployment from a GitHub repo with upload progress (uses XMLHttpRequest) */
export function createZipDeploymentWithProgress(
  params: {
    file: File;
    outputDirectory: string;
    buildCommand: string;
    friendCode?: string;
  },
  onProgress: (percent: number) => void
): Promise<Deployment> {
  return new Promise((resolve, reject) => {
    const formData = new FormData();
    formData.append("name", extractNameFromFilename(params.file.name));
    formData.append("source_type", "zip");
    formData.append("file", params.file);
    formData.append("output_directory", params.outputDirectory || ".");
    formData.append("build_command", params.buildCommand || "");
    if (params.friendCode) {
      formData.append("friend_code", params.friendCode);
    }

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${import.meta.env.VITE_API_BASE_URL || "http://localhost:8080"}/api/deployments`);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText));
      } else {
        try {
          const body = JSON.parse(xhr.responseText);
          reject(new Error(body.error || body.message || `Request failed with status ${xhr.status}`));
        } catch {
          reject(new Error(`Request failed with status ${xhr.status}`));
        }
      }
    };

    xhr.onerror = () => reject(new Error("Could not connect to the server. Please try again."));
    xhr.send(formData);
  });
}

