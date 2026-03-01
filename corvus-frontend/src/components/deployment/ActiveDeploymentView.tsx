import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import StatusBadge from "./StatusBadge";
import LiveUrlDisplay from "./LiveUrlDisplay";
import CountdownTimer from "./CountdownTimer";
import DeleteConfirmDialog from "./DeleteConfirmDialog";
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

/** Live deployment card shown on the landing page */
export default function ActiveDeploymentView({
  deployment,
  onDeleted,
  onRedeployStarted,
  onExpired,
}: ActiveDeploymentViewProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isRedeploying, setIsRedeploying] = useState(false);
  const { addToast } = useToast();
  const navigate = useNavigate();

  const expiresAt = new Date(
    new Date(deployment.created_at).getTime() + DEFAULT_TTL_MS
  );

  const handleDelete = useCallback(async () => {
    setIsDeleting(true);
    try {
      await deleteDeployment(deployment.id);
      addToast("Deployment deleted", "success");
      onDeleted();
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to delete deployment",
        "error"
      );
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  }, [deployment.id, onDeleted, addToast]);

  const handleRedeploy = useCallback(async () => {
    setIsRedeploying(true);
    try {
      const newDeployment = await redeployDeployment(deployment.id);
      onRedeployStarted(newDeployment);
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to redeploy",
        "error"
      );
    } finally {
      setIsRedeploying(false);
    }
  }, [deployment.id, onRedeployStarted, addToast]);

  return (
    <div className="bg-white border border-gray-200 rounded-xl p-6 max-w-md mx-auto text-center">
      <div className="mb-3">
        <StatusBadge status="live" />
      </div>

      <p className="text-lg font-semibold mb-1">Your site is live</p>
      <p className="text-sm text-gray-500 mb-4">"{deployment.name}"</p>

      {deployment.url && (
        <div className="mb-4 flex justify-center">
          <LiveUrlDisplay url={deployment.url} />
        </div>
      )}

      <div className="flex justify-center gap-3 mb-4">
        {deployment.url && (
          <button
            onClick={() => window.open(deployment.url, "_blank")}
            className="px-4 py-2 bg-black text-white rounded-lg text-sm hover:bg-gray-800 cursor-pointer"
          >
            Open Site
          </button>
        )}
      </div>

      <div className="mb-4">
        <CountdownTimer expiresAt={expiresAt} onExpired={onExpired} />
      </div>

      <div className="flex justify-center gap-3">
        <button
          onClick={() => navigate(`/d/${deployment.id}`)}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:border-black cursor-pointer"
        >
          View Details
        </button>
        <button
          onClick={handleRedeploy}
          disabled={isRedeploying}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:border-black cursor-pointer disabled:opacity-50"
        >
          {isRedeploying ? "Redeploying..." : "Redeploy"}
        </button>
        <button
          onClick={() => setShowDeleteDialog(true)}
          className="px-3 py-1.5 text-sm text-red-600 border border-red-200 rounded-lg hover:border-red-400 cursor-pointer"
        >
          Delete
        </button>
      </div>

      <DeleteConfirmDialog
        isOpen={showDeleteDialog}
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteDialog(false)}
        isDeleting={isDeleting}
      />
    </div>
  );
}

