import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import StatusBadge from "./StatusBadge";
import LiveUrlDisplay from "./LiveUrlDisplay";
import CountdownTimer from "./CountdownTimer";
import DeleteConfirmDialog from "./DeleteConfirmDialog";
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

/** Full detail card for the deployment viewer page */
export default function DeploymentDetailCard({
  deployment,
  onDeleted,
  onRedeployStarted,
  onExpired,
}: DeploymentDetailCardProps) {
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
      navigate("/");
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : "Failed to delete deployment",
        "error"
      );
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  }, [deployment.id, onDeleted, addToast, navigate]);

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
    <div className="bg-white border border-gray-200 rounded-xl p-6 max-w-lg mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">{deployment.name}</h2>
        <StatusBadge status={deployment.status} />
      </div>

      {/* Live URL */}
      {deployment.url && deployment.status === "live" && (
        <div className="mb-4">
          <LiveUrlDisplay url={deployment.url} />
        </div>
      )}

      {/* Countdown */}
      {deployment.status === "live" && (
        <div className="mb-4">
          <CountdownTimer expiresAt={expiresAt} onExpired={onExpired || (() => {})} />
        </div>
      )}

      {/* Action Buttons */}
      <div className="flex flex-wrap gap-2 mb-6">
        {deployment.url && deployment.status === "live" && (
          <button
            onClick={() => window.open(deployment.url, "_blank")}
            className="px-4 py-2 bg-black text-white rounded-lg text-sm hover:bg-gray-800 cursor-pointer"
          >
            Open Site
          </button>
        )}
        <button
          onClick={handleRedeploy}
          disabled={isRedeploying}
          className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:border-black cursor-pointer disabled:opacity-50"
        >
          {isRedeploying ? "Redeploying..." : "Redeploy"}
        </button>
        <button
          onClick={() => setShowDeleteDialog(true)}
          className="px-4 py-2 text-sm text-red-600 border border-red-200 rounded-lg hover:border-red-400 cursor-pointer"
        >
          Delete
        </button>
      </div>

      {/* Metadata */}
      <div className="border-t border-gray-200 pt-4 space-y-2 text-sm text-gray-500">
        <div className="flex justify-between">
          <span>Source</span>
          <span className="text-black">
            {deployment.source_type === "github" ? "GitHub" : "Zip Upload"}
          </span>
        </div>
        {deployment.github_url && (
          <div className="flex justify-between">
            <span>Repository</span>
            <a
              href={deployment.github_url.replace(/\.git$/, "")}
              target="_blank"
              rel="noopener noreferrer"
              className="text-black underline truncate max-w-[200px]"
            >
              {deployment.github_url.replace(/\.git$/, "").split("/").slice(-2).join("/")}
            </a>
          </div>
        )}
        <div className="flex justify-between">
          <span>Branch</span>
          <span className="text-black">{deployment.branch}</span>
        </div>
        <div className="flex justify-between">
          <span>Created</span>
          <span className="text-black">{formatTimestamp(deployment.created_at)}</span>
        </div>
        <div className="flex justify-between">
          <span>Updated</span>
          <span className="text-black">{formatTimestamp(deployment.updated_at)}</span>
        </div>
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

