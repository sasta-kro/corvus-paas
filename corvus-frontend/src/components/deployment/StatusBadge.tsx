import type { DeploymentStatus } from "../../types/deployment";

interface StatusBadgeProps {
  status: DeploymentStatus;
}

/** Small pill/badge showing deployment status */
export default function StatusBadge({ status }: StatusBadgeProps) {
  const styles = {
    deploying: "bg-gray-100 text-gray-700 border-gray-300",
    live: "bg-green-50 text-green-700 border-green-300",
    failed: "bg-red-50 text-red-700 border-red-300",
  };

  const labels = {
    deploying: "Deploying...",
    live: "Live",
    failed: "Failed",
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${styles[status]}`}
    >
      {status === "live" && <span className="mr-1">🟢</span>}
      {labels[status]}
    </span>
  );
}

