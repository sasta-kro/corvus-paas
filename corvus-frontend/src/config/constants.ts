import type { DeployPreset } from "../types/deployment";

// API
export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

// Deployment TTL
export const DEFAULT_TTL_MINUTES = 15;
export const DEFAULT_TTL_MS = DEFAULT_TTL_MINUTES * 60 * 1000;
export const EXTENDED_TTL_MINUTES = 60;
export const EXTENDED_TTL_MS = EXTENDED_TTL_MINUTES * 60 * 1000;

// Polling
export const POLL_INTERVAL_MS = 2000;
export const POLL_TIMEOUT_MS = 120000;

// Upload limits
export const MAX_FILE_SIZE_BYTES = 50 * 1024 * 1024;
export const MAX_FILE_SIZE_DISPLAY = "50MB";

// "Your Message" preset
export const MAX_MESSAGE_LENGTH = 100;

// localStorage keys
export const STORAGE_KEY_ACTIVE_DEPLOYMENT = "corvus_active_deployment";
export const STORAGE_KEY_FRIEND_CODE = "corvus_friend_code";

// Progress step timing (simulated)
export const STEP_DELAY_SOURCE_RECEIVED_MS = 2000;
export const STEP_DELAY_BUILDING_MS = 4000;
export const STEP_DELAY_STARTING_MS = 6000;

// Preset configurations
export const DEPLOY_PRESETS: DeployPreset[] = [
  {
    id: "vite-starter",
    name: "Vite Starter",
    description: "A minimal Vite app. Deploys in seconds.",
    icon: "⚡",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-vite-starter.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "react-app",
    name: "React App",
    description: "A React + Vite template.",
    icon: "⚛️",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-react-app.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "about-corvus",
    name: "About Corvus",
    description: "A custom page about this platform.",
    icon: "🐦",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-about.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: false,
  },
  {
    id: "your-message",
    name: "Your Message",
    description: "Create a page with your custom message.",
    icon: "✍️",
    githubUrl: "https://github.com/sasta-kro/corvus-preset-message.git",
    branch: "main",
    buildCommand: "npm ci && npm run build",
    outputDirectory: "dist",
    requiresTextInput: true,
  },
];

