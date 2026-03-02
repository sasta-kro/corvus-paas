/**
 * Deployment API — real implementations using the backend at API_BASE_URL.
 * All endpoints go through the shared client in client.ts.
 */
import { apiGet, apiPost, apiDelete, apiPostFormData } from "./client";
import { API_BASE_URL } from "../config/constants";
import type { Deployment } from "../types/deployment";
import { extractNameFromFilename } from "../lib/utils";

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
  formData.append("build_command", params.buildCommand);
  formData.append("output_directory", params.outputDirectory || ".");
  formData.append("branch", "main");
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
  formData.append("github_url", params.githubUrl);
  formData.append("branch", params.branch || "main");
  formData.append("build_command", params.buildCommand);
  formData.append("output_directory", params.outputDirectory || "dist");
  if (params.environmentVariables && Object.keys(params.environmentVariables).length > 0) {
    formData.append("environment_variables", JSON.stringify(params.environmentVariables));
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

/** Validates a friend code against the backend */
export async function validateFriendCode(
  code: string
): Promise<{ valid: boolean }> {
  return apiGet<{ valid: boolean }>(`/api/validate-code?code=${encodeURIComponent(code)}`);
}

/** Creates a deployment from a zip with upload progress tracking via XMLHttpRequest */
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
    formData.append("build_command", params.buildCommand);
    formData.append("output_directory", params.outputDirectory || ".");
    formData.append("branch", "main");
    if (params.friendCode) {
      formData.append("friend_code", params.friendCode);
    }

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API_BASE_URL}/api/deployments`);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText));
        } catch {
          reject(new Error("Failed to parse response"));
        }
      } else {
        try {
          const body = JSON.parse(xhr.responseText);
          reject(new Error(body.error || body.message || `Request failed with status ${xhr.status}`));
        } catch {
          reject(new Error(`Request failed with status ${xhr.status}`));
        }
      }
    };

    xhr.onerror = () => reject(new Error("Network error"));
    xhr.send(formData);
  });
}
