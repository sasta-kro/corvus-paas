/** Formats bytes into human-readable string */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const size = bytes / Math.pow(1024, i);
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

/** Extracts a deployment name from a zip filename */
export function extractNameFromFilename(filename: string): string {
  return filename.replace(/\.zip$/i, "");
}

/** Extracts a deployment name from a GitHub URL */
export function extractNameFromGithubUrl(url: string): string {
  const cleaned = url.replace(/\.git$/, "").replace(/\/$/, "");
  const parts = cleaned.split("/");
  return parts[parts.length - 1] || "deployment";
}

/** Ensures a GitHub URL ends with .git */
export function normalizeGithubUrl(url: string): string {
  const trimmed = url.trim().replace(/\/$/, "");
  if (trimmed.endsWith(".git")) return trimmed;
  return trimmed + ".git";
}

/** Validates that a string is a valid public GitHub repo URL */
export function isValidGithubUrl(url: string): boolean {
  try {
    const parsed = new URL(url.trim());
    return (
      parsed.hostname === "github.com" &&
      parsed.pathname.split("/").filter(Boolean).length >= 2
    );
  } catch {
    return false;
  }
}

/** Formats a countdown time as MM:SS */
export function formatCountdown(minutes: number, seconds: number): string {
  const m = String(minutes).padStart(2, "0");
  const s = String(seconds).padStart(2, "0");
  return `${m}:${s}`;
}

/** Formats an ISO timestamp to a human-readable string */
export function formatTimestamp(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
}

