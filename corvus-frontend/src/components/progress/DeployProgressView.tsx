import { useState, useEffect, useRef, useCallback } from "react";
import ProgressStep from "./ProgressStep";
import InkSplatter from "../shared/InkSplatter";
import { useDeploymentPolling } from "../../hooks/useDeploymentPolling";
import {
  STEP_DELAY_SOURCE_RECEIVED_MS,
  STEP_DELAY_BUILDING_MS,
  STEP_DELAY_STARTING_MS,
  POLL_TIMEOUT_MS,
} from "../../config/constants";
import type { Deployment } from "../../types/deployment";
import { deleteDeployment } from "../../api/deployments";

type StepStatus = "completed" | "in-progress" | "pending" | "failed";

const STEP_LABELS = [
  "Nest prepared",
  "Gathering materials",
  "Building the nest...",
  "Spreading wings",
  "Taking flight",
];

interface DeployProgressViewProps {
  deployment: Deployment;
  onComplete: (deployment: Deployment) => void;
  onFailed: () => void;
  onCancel: () => void;
}

export default function DeployProgressView({
  deployment, onComplete, onFailed, onCancel,
}: DeployProgressViewProps) {
  const { deployment: polledDeployment } = useDeploymentPolling(deployment.id);
  const [stepStatuses, setStepStatuses] = useState<StepStatus[]>([
    "completed", "in-progress", "pending", "pending", "pending",
  ]);
  const [isTimedOut, setIsTimedOut] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const startTimeRef = useRef(Date.now());
  const completedRef = useRef(false);
  const timeoutExtendedRef = useRef(false);

  useEffect(() => {
    const timers: ReturnType<typeof setTimeout>[] = [];
    timers.push(setTimeout(() => {
      setStepStatuses((prev) => {
        const next = [...prev];
        if (next[1] === "in-progress" || next[1] === "pending") { next[1] = "completed"; if (next[2] === "pending") next[2] = "in-progress"; }
        return next;
      });
    }, STEP_DELAY_SOURCE_RECEIVED_MS));
    timers.push(setTimeout(() => {
      setStepStatuses((prev) => { const next = [...prev]; if (next[2] === "pending") next[2] = "in-progress"; return next; });
    }, STEP_DELAY_BUILDING_MS));
    timers.push(setTimeout(() => {
      setStepStatuses((prev) => {
        const next = [...prev];
        if (next[2] === "in-progress") { next[2] = "completed"; if (next[3] === "pending") next[3] = "in-progress"; }
        return next;
      });
    }, STEP_DELAY_STARTING_MS));
    return () => timers.forEach(clearTimeout);
  }, []);

  useEffect(() => {
    const checkTimeout = setInterval(() => {
      if (Date.now() - startTimeRef.current > POLL_TIMEOUT_MS && !completedRef.current) setIsTimedOut(true);
    }, 1000);
    return () => clearInterval(checkTimeout);
  }, []);

  useEffect(() => {
    if (!polledDeployment || completedRef.current) return;
    if (polledDeployment.status === "live") {
      completedRef.current = true;
      setStepStatuses(["completed", "completed", "completed", "completed", "completed"]);
      setTimeout(() => onComplete(polledDeployment), 500);
    } else if (polledDeployment.status === "failed") {
      completedRef.current = true;
      setStepStatuses((prev) => {
        const next = [...prev]; const idx = next.findIndex((s) => s === "in-progress");
        if (idx !== -1) next[idx] = "failed"; return next;
      });
    }
  }, [polledDeployment, onComplete]);

  const handleKeepWaiting = useCallback(() => {
    startTimeRef.current = Date.now(); setIsTimedOut(false); timeoutExtendedRef.current = true;
  }, []);

  const handleCancel = useCallback(async () => {
    setIsCancelling(true);
    try { await deleteDeployment(deployment.id); } catch { /* ignore */ }
    onCancel();
  }, [deployment.id, onCancel]);

  const isFailed = polledDeployment?.status === "failed";

  return (
    <div className="ink-card torn-edge-2 max-w-md mx-auto relative" style={{ zIndex: 10 }}>
      {/* Decorative ink mark */}
      <InkSplatter variant={0} size={50} style={{ top: -8, right: -8, opacity: 0.10 }} />

      <h3 className="font-brush text-lg mb-5" style={{ color: "var(--sumi)" }}>
        Deploying "{deployment.name}"
      </h3>

      <div>
        {STEP_LABELS.map((label, i) => (
          <ProgressStep key={label} label={label} status={stepStatuses[i]} />
        ))}
      </div>

      {isFailed && (
        <div className="mt-6 text-center">
          <p style={{ color: "var(--vermillion)", fontSize: "0.9rem", marginBottom: "0.75rem" }}>
            The crow fell. Check the branch, build command, and output directory, then try again.
          </p>
          <button onClick={onFailed} className="ink-btn">Try Again</button>
        </div>
      )}

      {isTimedOut && !isFailed && !completedRef.current && (
        <div className="mt-6 text-center">
          <p style={{ color: "var(--sumi-light)", fontSize: "0.9rem", marginBottom: "0.75rem", fontStyle: "italic" }}>
            The crow is still circling. It may yet land.
          </p>
          <div className="flex justify-center gap-3">
            <button onClick={handleKeepWaiting} className="ink-btn">Keep Waiting</button>
            <button onClick={handleCancel} disabled={isCancelling} className="ink-btn-outline">
              {isCancelling ? "Cancelling..." : "Cancel"}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
