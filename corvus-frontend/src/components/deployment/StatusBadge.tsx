import type { DeploymentStatus } from "../../types/deployment";

interface StatusBadgeProps {
  status: DeploymentStatus;
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const config: Record<DeploymentStatus, { bg: string; color: string; border: string; label: string; dot?: string }> = {
    deploying: {
      bg: "var(--paper-warm)",
      color: "var(--sumi-light)",
      border: "var(--sumi-ghost)",
      label: "Deploying...",
    },
    live: {
      bg: "var(--leaf-bg)",
      color: "var(--leaf)",
      border: "var(--leaf)",
      label: "Live",
      dot: "var(--leaf)",
    },
    failed: {
      bg: "var(--vermillion-bg)",
      color: "var(--vermillion)",
      border: "var(--vermillion)",
      label: "Failed",
    },
  };

  const c = config[status];

  return (
    <span
      className="inline-flex items-center gap-1.5 px-3 py-1 text-xs"
      style={{
        background: c.bg,
        color: c.color,
        border: `1.5px solid ${c.border}`,
        borderRadius: "1px 3px 2px 1px",
        fontFamily: '"EB Garamond", serif',
        fontWeight: 700,
        letterSpacing: "0.04em",
        textTransform: "uppercase",
        fontSize: "0.7rem",
      }}
    >
      {c.dot && (
        <span style={{ width: 6, height: 6, borderRadius: "50%", background: c.dot, display: "inline-block" }} />
      )}
      {c.label}
    </span>
  );
}
