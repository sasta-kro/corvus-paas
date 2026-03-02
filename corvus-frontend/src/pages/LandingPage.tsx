import { useState, useEffect, useCallback } from "react";
import HeroSection from "../components/layout/HeroSection";
import DeployPanel from "../components/deploy/DeployPanel";
import DeployProgressView from "../components/progress/DeployProgressView";
import ActiveDeploymentView from "../components/deployment/ActiveDeploymentView";
import InkSpinner from "../components/shared/InkSpinner";
import { useActiveDeployment } from "../hooks/useActiveDeployment";
import { getDeployment } from "../api/deployments";
import { useToast } from "../components/shared/Toast";
import type { Deployment } from "../types/deployment";

type ViewState = "checking" | "deploy" | "progress" | "active";

export default function LandingPage() {
  const [viewState, setViewState] = useState<ViewState>("checking");
  const [deployment, setDeployment] = useState<Deployment | null>(null);
  const { activeDeployment, setActiveDeployment, clearActiveDeployment } = useActiveDeployment();
  const { addToast } = useToast();

  useEffect(() => {
    if (!activeDeployment) { setViewState("deploy"); return; }
    let cancelled = false;
    const check = async () => {
      try {
        const data = await getDeployment(activeDeployment.id);
        if (cancelled) return;
        setDeployment(data);
        if (data.status === "live") setViewState("active");
        else if (data.status === "deploying") setViewState("progress");
        else if (data.status === "failed") { clearActiveDeployment(); setViewState("deploy"); addToast("Previous deployment failed", "error"); }
      } catch { if (!cancelled) { clearActiveDeployment(); setViewState("deploy"); } }
    };
    check();
    return () => { cancelled = true; };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleDeployStarted = useCallback((dep: Deployment) => {
    setDeployment(dep); setActiveDeployment(dep.id, dep.slug); setViewState("progress");
  }, [setActiveDeployment]);

  const handleProgressComplete = useCallback((dep: Deployment) => {
    setDeployment(dep); setViewState("active");
  }, []);

  const handleProgressFailed = useCallback(() => {
    clearActiveDeployment(); setDeployment(null); setViewState("deploy");
  }, [clearActiveDeployment]);

  const handleProgressCancel = useCallback(() => {
    clearActiveDeployment(); setDeployment(null); setViewState("deploy");
  }, [clearActiveDeployment]);

  const handleDeleted = useCallback(() => {
    clearActiveDeployment(); setDeployment(null); setViewState("deploy");
  }, [clearActiveDeployment]);

  const handleRedeployStarted = useCallback((dep: Deployment) => {
    setDeployment(dep); setActiveDeployment(dep.id, dep.slug); setViewState("progress");
  }, [setActiveDeployment]);

  const handleExpired = useCallback(() => {
    if (!deployment) return;
    getDeployment(deployment.id)
      .then((data) => { if (data.status === "live") setTimeout(handleExpired, 5000); else { clearActiveDeployment(); setDeployment(null); setViewState("deploy"); } })
      .catch(() => { clearActiveDeployment(); setDeployment(null); setViewState("deploy"); });
  }, [deployment, clearActiveDeployment]);

  return (
    <div className="flex-1 relative" style={{ zIndex: 10 }}>
      <HeroSection />

      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-4 watercolor-wash">
        {viewState === "checking" && (
          <div className="text-center py-16">
            <InkSpinner size="lg" />
            <p style={{ color: "var(--sumi-wash)", fontSize: "0.9rem", marginTop: "1rem", fontStyle: "italic" }}>
              Checking status...
            </p>
          </div>
        )}

        {viewState === "deploy" && <DeployPanel onDeployStarted={handleDeployStarted} />}

        {viewState === "progress" && deployment && (
          <DeployProgressView deployment={deployment} onComplete={handleProgressComplete}
            onFailed={handleProgressFailed} onCancel={handleProgressCancel} />
        )}

        {viewState === "active" && deployment && (
          <ActiveDeploymentView deployment={deployment} onDeleted={handleDeleted}
            onRedeployStarted={handleRedeployStarted} onExpired={handleExpired} />
        )}
      </div>
    </div>
  );
}
