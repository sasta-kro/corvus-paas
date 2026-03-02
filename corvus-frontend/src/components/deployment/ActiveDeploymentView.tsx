import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import StatusBadge from "./StatusBadge";
import LiveUrlDisplay from "./LiveUrlDisplay";
import CountdownTimer from "./CountdownTimer";
import DeleteConfirmDialog from "./DeleteConfirmDialog";
import InkSplatter from "../shared/InkSplatter";
import { deleteDeployment, redeployDeployment } from "../../api/deployments";
import { DEFAULT_TTL_MS } from "../../config/constants";
import { useToast } from "../shared/Toast";
import type { Deployment } from "../../types/deployment";

interface ActiveDeploymentViewProps {
  deployment: Deployment;
  onDeleted: () => void;
  onRedeployStarted: (deployment: Deployment) => void;
  onExpired: () => void;
}

export default function ActiveDeploymentView({
  deployment, onDeleted, onRedeployStarted, onExpired,
}: ActiveDeploymentViewProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isRedeploying, setIsRedeploying] = useState(false);
  const { addToast } = useToast();
  const navigate = useNavigate();

  // Use expires_at from backend if available, otherwise fall back to client-side calculation
  const expiresAt = deployment.expires_at
    ? new Date(deployment.expires_at)
    : new Date(new Date(deployment.created_at).getTime() + DEFAULT_TTL_MS);

  const handleDelete = useCallback(async () => {
    setIsDeleting(true);
    try { await deleteDeployment(deployment.id); addToast("Deployment deleted", "success"); onDeleted(); }
    catch (err) { addToast(err instanceof Error ? err.message : "Failed to delete deployment", "error"); }
    finally { setIsDeleting(false); setShowDeleteDialog(false); }
  }, [deployment.id, onDeleted, addToast]);

  const handleRedeploy = useCallback(async () => {
    setIsRedeploying(true);
    try { const d = await redeployDeployment(deployment.id); onRedeployStarted(d); }
    catch (err) { addToast(err instanceof Error ? err.message : "Failed to redeploy", "error"); }
    finally { setIsRedeploying(false); }
  }, [deployment.id, onRedeployStarted, addToast]);

  return (
    <>
    <div className="ink-card torn-edge-3 max-w-md mx-auto text-center relative" style={{ zIndex: 10 }}>
      <InkSplatter variant={2} size={60} style={{ top: -10, left: -10, opacity: 0.09 }} />

      <div className="mb-4">
        <StatusBadge status="live" />
      </div>

      <p className="font-brush text-xl mb-1" style={{ color: "var(--sumi)" }}>
        Your site has taken flight
      </p>
      <p style={{ color: "var(--sumi-light)", fontSize: "0.9rem", marginBottom: "1.25rem", fontStyle: "italic" }}>
        "{deployment.name}"
      </p>

      {deployment.url && (
        <div className="mb-5 flex justify-center">
          <LiveUrlDisplay url={deployment.url} />
        </div>
      )}

      <div className="flex justify-center gap-3 mb-5">
        {deployment.url && (
          <button onClick={() => window.open(deployment.url, "_blank")} className="ink-btn">
            Open Site
          </button>
        )}
      </div>

      <div className="mb-5">
        <CountdownTimer expiresAt={expiresAt} onExpired={onExpired} />
      </div>

      <div className="flex justify-center gap-3">
        <button onClick={() => navigate(`/d/${deployment.id}`)} className="ink-btn-outline">View Details</button>
        <button onClick={handleRedeploy} disabled={isRedeploying} className="ink-btn-outline">
          {isRedeploying ? "Redeploying..." : "Redeploy"}
        </button>
        <button onClick={() => setShowDeleteDialog(true)} className="ink-btn-danger">Delete</button>
      </div>

    </div>

    <DeleteConfirmDialog isOpen={showDeleteDialog} onConfirm={handleDelete}
      onCancel={() => setShowDeleteDialog(false)} isDeleting={isDeleting} />
  </>
  );
}
