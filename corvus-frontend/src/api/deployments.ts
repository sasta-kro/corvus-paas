/**
 * Mock Deployment API — returns fake data with simulated delays
 * so the frontend can be tested without a running backend.
 *
 * Original real implementations are preserved in deployments.real.ts
 * (or just revert this file from git to restore them).
 */
import type { Deployment, SourceType } from "../types/deployment";
import { extractNameFromFilename } from "../lib/utils";

// ---------------------------------------------------------------------------
// In-memory store for mock deployments
// ---------------------------------------------------------------------------
const mockStore = new Map<string, { deployment: Deployment; pollCount: number }>();

const POLL_THRESHOLD = 4; // after this many getDeployment() calls, status → "live"

function randomId(): string {
  return Math.random().toString(36).slice(2, 10);
}

function makeMockDeployment(overrides: Partial<Deployment> & { name: string; source_type: SourceType }): Deployment {
  const id = "mock-deploy-" + randomId();
  const slug = "mock-site-" + randomId();
  const now = new Date().toISOString();

  const deployment: Deployment = {
    id,
    slug,
    name: overrides.name,
    source_type: overrides.source_type,
    github_url: overrides.github_url,
    branch: overrides.branch ?? "main",
    build_command: overrides.build_command ?? "npm ci && npm run build",
    output_directory: overrides.output_directory ?? "dist",
    url: undefined,
    auto_deploy: false,
    created_at: now,
    updated_at: now,
    ...overrides,
    // Always start as deploying regardless of overrides
    status: "deploying",
  };

  mockStore.set(id, { deployment, pollCount: 0 });
  return deployment;
}

function delay(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

// ---------------------------------------------------------------------------
// Mock API functions (same signatures as the real ones)
// ---------------------------------------------------------------------------

/** Creates a new deployment from a zip file upload */
export async function createZipDeployment(params: {
  file: File;
  outputDirectory: string;
  buildCommand: string;
  friendCode?: string;
}): Promise<Deployment> {
  await delay(800);
  return makeMockDeployment({
    name: extractNameFromFilename(params.file.name),
    source_type: "zip",
    build_command: params.buildCommand || "npm ci && npm run build",
    output_directory: params.outputDirectory || ".",
  });
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
  await delay(800);
  return makeMockDeployment({
    name: params.name,
    source_type: "github",
    github_url: params.githubUrl,
    branch: params.branch || "main",
    build_command: params.buildCommand,
    output_directory: params.outputDirectory || "dist",
  });
}

/** Fetches a single deployment by ID — simulates build completing after a few polls */
export async function getDeployment(id: string): Promise<Deployment> {
  await delay(300);

  const entry = mockStore.get(id);
  if (!entry) {
    // If we don't know about this id (e.g. recovered from localStorage),
    // create a stub so the UI has something to render.
    const stub = makeMockDeployment({
      name: "recovered-site",
      source_type: "zip",
    });
    // Override the id/slug to match what was requested
    stub.id = id;
    mockStore.set(id, { deployment: stub, pollCount: 0 });
    return { ...stub };
  }

  entry.pollCount++;

  if (entry.pollCount >= POLL_THRESHOLD && entry.deployment.status === "deploying") {
    entry.deployment.status = "live";
    entry.deployment.url = `https://${entry.deployment.slug}.corvus.example.com`;
    entry.deployment.updated_at = new Date().toISOString();
  }

  return { ...entry.deployment };
}

/** Deletes a deployment by ID */
export async function deleteDeployment(id: string): Promise<void> {
  await delay(500);
  mockStore.delete(id);
}

/** Triggers a redeploy for an existing deployment */
export async function redeployDeployment(id: string): Promise<Deployment> {
  await delay(600);

  const entry = mockStore.get(id);
  const baseName = entry?.deployment.name ?? "redeployed-site";
  const sourceType = entry?.deployment.source_type ?? "zip";

  // Remove old entry
  mockStore.delete(id);

  return makeMockDeployment({
    name: baseName,
    source_type: sourceType,
    github_url: entry?.deployment.github_url,
    branch: entry?.deployment.branch ?? "main",
    build_command: entry?.deployment.build_command ?? "npm ci && npm run build",
    output_directory: entry?.deployment.output_directory ?? "dist",
  });
}

/** Validates a friend code — accepts any non-empty string */
export async function validateFriendCode(
  code: string
): Promise<{ valid: boolean }> {
  await delay(300);
  return { valid: code.trim().length > 0 };
}

/** Creates a deployment from a zip with simulated upload progress */
export function createZipDeploymentWithProgress(
  params: {
    file: File;
    outputDirectory: string;
    buildCommand: string;
    friendCode?: string;
  },
  onProgress: (percent: number) => void
): Promise<Deployment> {
  return new Promise((resolve) => {
    let progress = 0;
    const interval = setInterval(() => {
      progress += Math.floor(Math.random() * 15) + 5;
      if (progress >= 100) {
        progress = 100;
        clearInterval(interval);
        onProgress(100);

        // After "upload" completes, create the mock deployment
        delay(200).then(() => {
          resolve(
            makeMockDeployment({
              name: extractNameFromFilename(params.file.name),
              source_type: "zip",
              build_command: params.buildCommand || "npm ci && npm run build",
              output_directory: params.outputDirectory || ".",
            })
          );
        });
        return;
      }
      onProgress(progress);
    }, 150);
  });
}
