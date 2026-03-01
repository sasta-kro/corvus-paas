export type DeploymentStatus = "deploying" | "live" | "failed";
export type SourceType = "zip" | "github";

export interface Deployment {
  id: string;
  slug: string;
  name: string;
  source_type: SourceType;
  github_url?: string;
  branch: string;
  build_command: string;
  output_directory: string;
  environment_variables?: string;
  status: DeploymentStatus;
  url?: string;
  webhook_secret?: string;
  auto_deploy: boolean;
  created_at: string;
  updated_at: string;
}

export interface DeployPreset {
  id: string;
  name: string;
  description: string;
  icon: string;
  githubUrl: string;
  branch: string;
  buildCommand: string;
  outputDirectory: string;
  requiresTextInput: boolean;
  environmentVariables?: Record<string, string>;
}

export interface ActiveDeploymentSession {
  id: string;
  slug: string;
}

