import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import StatusBadge from "./StatusBadge";
import LiveUrlDisplay from "./LiveUrlDisplay";
import CountdownTimer from "./CountdownTimer";
import DeleteConfirmDialog from "./DeleteConfirmDialog";
import InkSplatter from "../shared/InkSplatter";
import { deleteDeployment, redeployDeployment } from "../../api/deployments";
import { DEFAULT_TTL_MS } from "../../config/constants";
import { formatTimestamp } from "../../lib/utils";
import { useToast } from "../shared/Toast";
import type { Deployment } from "../../types/deployment";

interface DeploymentDetailCardProps {
  deployment: Deployment;
  onDeleted: () => void;
  onRedeployStarted: (deployment: Deployment) => void;
  onExpired?: () => void;
}

export default function DeploymentDetailCard({
  deployment, onDeleted, onRedeployStarted, onExpired,
}: DeploymentDetailCardProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isRedeploying, setIsRedeploying] = useState(false);
  const { addToast } = useToast();
  const navigate = useNavigate();

  const expiresAt = deployment.expires_at
    ? new Date(deployment.expires_at)
    : new Date(new Date(deployment.created_at).getTime() + DEFAULT_TTL_MS);

  const handleDelete = useCallback(async () => {
    setIsDeleting(true);
    try { await deleteDeployment(deployment.id); addToast("Deployment deleted", "success"); onDeleted(); navigate("/"); }
    catch (err) { addToast(err instanceof Error ? err.message : "Failed to delete deployment", "error"); }
    finally { setIsDeleting(false); setShowDeleteDialog(false); }
  }, [deployment.id, onDeleted, addToast, navigate]);

  const handleRedeploy = useCallback(async () => {
    setIsRedeploying(true);
    try { const d = await redeployDeployment(deployment.id); onRedeployStarted(d); }
    catch (err) { addToast(err instanceof Error ? err.message : "Failed to redeploy", "error"); }
    finally { setIsRedeploying(false); }
  }, [deployment.id, onRedeployStarted, addToast]);

  return (
    <>
    <div className="ink-card torn-edge-1 max-w-lg mx-auto relative" style={{ zIndex: 10 }}>
      <InkSplatter variant={1} size={45} style={{ top: -6, right: -6, opacity: 0.09 }} />

      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <h2 className="font-brush text-xl" style={{ color: "var(--sumi)" }}>{deployment.name}</h2>
        <StatusBadge status={deployment.status} />
      </div>

      {/* Live URL */}
      {deployment.url && deployment.status === "live" && (
        <div className="mb-5"><LiveUrlDisplay url={deployment.url} /></div>
      )}

      {/* Countdown */}
      {deployment.status === "live" && (
        <div className="mb-5"><CountdownTimer expiresAt={expiresAt} onExpired={onExpired || (() => {})} /></div>
      )}

      {/* Actions */}
      <div className="flex flex-wrap gap-3 mb-6">
        {deployment.url && deployment.status === "live" && (
          <button onClick={() => window.open(deployment.url, "_blank")} className="ink-btn">Open Site</button>
        )}
        <button onClick={handleRedeploy} disabled={isRedeploying} className="ink-btn-outline">
          {isRedeploying ? "Redeploying..." : "Redeploy"}
        </button>
        <button onClick={() => setShowDeleteDialog(true)} className="ink-btn-danger">Delete</button>
      </div>

      {/* Metadata */}
      <div className="brush-divider-thin mb-5" />
      <div className="space-y-3" style={{ fontSize: "0.9rem" }}>
        {[
          { label: "Source", value: deployment.source_type === "github" ? "GitHub" : "Zip Upload" },
          ...(deployment.github_url ? [{
            label: "Repository",
            value: deployment.github_url.replace(/\.git$/, "").split("/").slice(-2).join("/"),
            href: deployment.github_url.replace(/\.git$/, ""),
          }] : []),
          { label: "Branch", value: deployment.branch },
          { label: "Created", value: formatTimestamp(deployment.created_at) },
          { label: "Updated", value: formatTimestamp(deployment.updated_at) },
        ].map((row) => (
          <div key={row.label} className="flex justify-between">
            <span className="ink-label" style={{ marginBottom: 0, textTransform: "none", fontSize: "0.85rem" }}>
              {row.label}
            </span>
            {"href" in row && row.href ? (
              <a href={row.href} target="_blank" rel="noopener noreferrer"
                className="truncate max-w-[200px]"
                style={{ color: "var(--sumi)", textDecoration: "underline", textUnderlineOffset: "2px" }}>
                {row.value}
              </a>
            ) : (
              <span style={{ color: "var(--sumi)", fontWeight: 700 }}>{row.value}</span>
            )}
          </div>
        ))}
      </div>

    </div>

    <DeleteConfirmDialog isOpen={showDeleteDialog} onConfirm={handleDelete}
      onCancel={() => setShowDeleteDialog(false)} isDeleting={isDeleting} />
  </>
  );
}
