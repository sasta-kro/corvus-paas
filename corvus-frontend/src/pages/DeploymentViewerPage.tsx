import { useState, useEffect, useCallback } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import DeploymentDetailCard from "../components/deployment/DeploymentDetailCard";
import DeployProgressView from "../components/progress/DeployProgressView";
import InkSpinner from "../components/shared/InkSpinner";
import { useDeploymentPolling } from "../hooks/useDeploymentPolling";
import { useActiveDeployment } from "../hooks/useActiveDeployment";
import { getDeployment } from "../api/deployments";
import type { Deployment } from "../types/deployment";

export default function DeploymentViewerPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { clearActiveDeployment, setActiveDeployment } = useActiveDeployment();
  const { deployment, isLoading, isNotFound } = useDeploymentPolling(id || null, 5000, true);
  const [redeployedDeployment, setRedeployedDeployment] = useState<Deployment | null>(null);
  const [showProgress, setShowProgress] = useState(false);

  useEffect(() => { if (deployment?.status === "deploying") setShowProgress(true); }, [deployment?.status]);

  const handleDeleted = useCallback(() => { clearActiveDeployment(); navigate("/"); }, [clearActiveDeployment, navigate]);

  const handleRedeployStarted = useCallback((dep: Deployment) => {
    setRedeployedDeployment(dep); setActiveDeployment(dep.id, dep.slug); setShowProgress(true);
  }, [setActiveDeployment]);

  const handleProgressComplete = useCallback((dep: Deployment) => {
    setRedeployedDeployment(null); setShowProgress(false); void dep;
  }, []);

  const handleProgressFailed = useCallback(() => { setRedeployedDeployment(null); setShowProgress(false); }, []);

  const handleProgressCancel = useCallback(() => {
    setRedeployedDeployment(null); setShowProgress(false); clearActiveDeployment(); navigate("/");
  }, [clearActiveDeployment, navigate]);

  const handleExpired = useCallback(() => {
    if (!id) return;
    getDeployment(id)
      .then((data) => { if (data.status === "live") setTimeout(handleExpired, 5000); else { clearActiveDeployment(); navigate("/"); } })
      .catch(() => { clearActiveDeployment(); navigate("/"); });
  }, [id, clearActiveDeployment, navigate]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center py-20 relative" style={{ zIndex: 10 }}>
        <div className="text-center">
          <InkSpinner size="lg" />
          <p style={{ color: "var(--sumi-wash)", fontSize: "0.9rem", marginTop: "1rem", fontStyle: "italic" }}>
            Loading deployment...
          </p>
        </div>
      </div>
    );
  }

  if (isNotFound || !deployment) {
    return (
      <div className="flex-1 flex items-center justify-center py-20 relative" style={{ zIndex: 10 }}>
        <div className="text-center">
          <h2 className="font-brush text-2xl mb-3" style={{ color: "var(--sumi)" }}>
            The crow has flown
          </h2>
          <p style={{ color: "var(--sumi-light)", marginBottom: "1.25rem", fontStyle: "italic" }}>
            This deployment may have expired or been deleted.
          </p>
          <Link to="/" className="ink-btn-outline" style={{ textDecoration: "none" }}>
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" style={{ display: "inline-block", verticalAlign: "middle", marginRight: "4px" }}><path d="M9 2L4 7L9 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" /></svg>Back to home
          </Link>
        </div>
      </div>
    );
  }

  const activeDeployment = redeployedDeployment || deployment;

  if (showProgress && activeDeployment.status === "deploying") {
    return (
      <div className="flex-1 py-12 relative" style={{ zIndex: 10 }}>
        <div className="max-w-3xl mx-auto px-4 sm:px-6">
          <DeployProgressView deployment={activeDeployment} onComplete={handleProgressComplete}
            onFailed={handleProgressFailed} onCancel={handleProgressCancel} />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 py-12 relative" style={{ zIndex: 10 }}>
      <div className="max-w-3xl mx-auto px-4 sm:px-6">
        <div className="mb-6">
          <Link to="/" className="transition-colors" style={{
            color: "var(--sumi-wash)", fontSize: "0.9rem", textDecoration: "none",
          }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "var(--sumi)")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "var(--sumi-wash)")}
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" style={{ display: "inline-block", verticalAlign: "middle", marginRight: "4px" }}><path d="M9 2L4 7L9 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" /></svg>Back to home
          </Link>
        </div>
        <DeploymentDetailCard deployment={deployment} onDeleted={handleDeleted}
          onRedeployStarted={handleRedeployStarted} onExpired={handleExpired} />
      </div>
    </div>
  );
}
