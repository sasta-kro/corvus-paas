import { useState, useEffect, useCallback } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import DeploymentDetailCard from "../components/deployment/DeploymentDetailCard";
import DeployProgressView from "../components/progress/DeployProgressView";
import { useDeploymentPolling } from "../hooks/useDeploymentPolling";
import { useActiveDeployment } from "../hooks/useActiveDeployment";
import { getDeployment } from "../api/deployments";
import type { Deployment } from "../types/deployment";

/** Deployment viewer page — shows full deployment details for /d/:id */
export default function DeploymentViewerPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { clearActiveDeployment, setActiveDeployment } = useActiveDeployment();
  const { deployment, isLoading, isNotFound } = useDeploymentPolling(
    id || null,
    5000,
    true
  );
  const [redeployedDeployment, setRedeployedDeployment] = useState<Deployment | null>(null);
  const [showProgress, setShowProgress] = useState(false);

  // If the deployment is deploying, show progress view
  useEffect(() => {
    if (deployment?.status === "deploying") {
      setShowProgress(true);
    }
  }, [deployment?.status]);

  const handleDeleted = useCallback(() => {
    clearActiveDeployment();
    navigate("/");
  }, [clearActiveDeployment, navigate]);

  const handleRedeployStarted = useCallback(
    (dep: Deployment) => {
      setRedeployedDeployment(dep);
      setActiveDeployment(dep.id, dep.slug);
      setShowProgress(true);
    },
    [setActiveDeployment]
  );

  const handleProgressComplete = useCallback((dep: Deployment) => {
    setRedeployedDeployment(null);
    setShowProgress(false);
    // The polling hook will pick up the new status
    void dep;
  }, []);

  const handleProgressFailed = useCallback(() => {
    setRedeployedDeployment(null);
    setShowProgress(false);
  }, []);

  const handleProgressCancel = useCallback(() => {
    setRedeployedDeployment(null);
    setShowProgress(false);
    clearActiveDeployment();
    navigate("/");
  }, [clearActiveDeployment, navigate]);

  const handleExpired = useCallback(() => {
    if (!id) return;
    getDeployment(id)
      .then((data) => {
        if (data.status === "live") {
          // Still alive, backend cleanup slightly delayed — re-check in 5s
          setTimeout(handleExpired, 5000);
        } else {
          clearActiveDeployment();
          navigate("/");
        }
      })
      .catch(() => {
        // 404 or error — deployment is gone
        clearActiveDeployment();
        navigate("/");
      });
  }, [id, clearActiveDeployment, navigate]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center py-20">
        <div className="text-center">
          <div className="inline-block animate-spin h-8 w-8 border-4 border-gray-200 border-t-black rounded-full" />
          <p className="text-sm text-gray-500 mt-3">Loading deployment...</p>
        </div>
      </div>
    );
  }

  if (isNotFound || !deployment) {
    return (
      <div className="flex-1 flex items-center justify-center py-20">
        <div className="text-center">
          <h2 className="text-2xl font-semibold mb-2">Deployment not found</h2>
          <p className="text-gray-500 mb-4">
            This deployment may have expired or been deleted.
          </p>
          <Link
            to="/"
            className="text-black underline hover:text-gray-600"
          >
            ← Back to home
          </Link>
        </div>
      </div>
    );
  }

  const activeDeployment = redeployedDeployment || deployment;

  if (showProgress && activeDeployment.status === "deploying") {
    return (
      <div className="flex-1 py-12">
        <div className="max-w-3xl mx-auto px-4 sm:px-6">
          <DeployProgressView
            deployment={activeDeployment}
            onComplete={handleProgressComplete}
            onFailed={handleProgressFailed}
            onCancel={handleProgressCancel}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 py-12">
      <div className="max-w-3xl mx-auto px-4 sm:px-6">
        <div className="mb-6">
          <Link
            to="/"
            className="text-sm text-gray-500 hover:text-black"
          >
            ← Back to home
          </Link>
        </div>
        <DeploymentDetailCard
          deployment={deployment}
          onDeleted={handleDeleted}
          onRedeployStarted={handleRedeployStarted}
          onExpired={handleExpired}
        />
      </div>
    </div>
  );
}

