import { useState, useEffect, useRef, useCallback } from "react";
import ProgressStep from "./ProgressStep";
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
  "Deployment created",
  "Source files received",
  "Building project...",
  "Starting server",
  "Going live",
];

interface DeployProgressViewProps {
  deployment: Deployment;
  onComplete: (deployment: Deployment) => void;
  onFailed: () => void;
  onCancel: () => void;
}

/** Progress steps container — polls deployment status and advances visual steps */
export default function DeployProgressView({
  deployment,
  onComplete,
  onFailed,
  onCancel,
}: DeployProgressViewProps) {
  const { deployment: polledDeployment } = useDeploymentPolling(deployment.id);
  const [stepStatuses, setStepStatuses] = useState<StepStatus[]>([
    "completed",
    "in-progress",
    "pending",
    "pending",
    "pending",
  ]);
  const [isTimedOut, setIsTimedOut] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const startTimeRef = useRef(Date.now());
  const completedRef = useRef(false);
  const timeoutExtendedRef = useRef(false);

  // Timer-based step advancement
  useEffect(() => {
    const timers: ReturnType<typeof setTimeout>[] = [];

    timers.push(
      setTimeout(() => {
        setStepStatuses((prev) => {
          const next = [...prev];
          if (next[1] === "in-progress" || next[1] === "pending") {
            next[1] = "completed";
            if (next[2] === "pending") next[2] = "in-progress";
          }
          return next;
        });
      }, STEP_DELAY_SOURCE_RECEIVED_MS)
    );

    timers.push(
      setTimeout(() => {
        setStepStatuses((prev) => {
          const next = [...prev];
          if (next[2] === "pending") next[2] = "in-progress";
          return next;
        });
      }, STEP_DELAY_BUILDING_MS)
    );

    timers.push(
      setTimeout(() => {
        setStepStatuses((prev) => {
          const next = [...prev];
          if (next[2] === "in-progress") {
            next[2] = "completed";
            if (next[3] === "pending") next[3] = "in-progress";
          }
          return next;
        });
      }, STEP_DELAY_STARTING_MS)
    );

    return () => timers.forEach(clearTimeout);
  }, []);

  // Timeout check
  useEffect(() => {
    const checkTimeout = setInterval(() => {
      const elapsed = Date.now() - startTimeRef.current;
      if (elapsed > POLL_TIMEOUT_MS && !completedRef.current) {
        setIsTimedOut(true);
      }
    }, 1000);
    return () => clearInterval(checkTimeout);
  }, []);

  // React to polled status changes
  useEffect(() => {
    if (!polledDeployment || completedRef.current) return;

    if (polledDeployment.status === "live") {
      completedRef.current = true;
      setStepStatuses(["completed", "completed", "completed", "completed", "completed"]);
      setTimeout(() => onComplete(polledDeployment), 500);
    } else if (polledDeployment.status === "failed") {
      completedRef.current = true;
      setStepStatuses((prev) => {
        const next = [...prev];
        // Find the in-progress step and mark it failed
        const inProgressIdx = next.findIndex((s) => s === "in-progress");
        if (inProgressIdx !== -1) {
          next[inProgressIdx] = "failed";
        }
        return next;
      });
    }
  }, [polledDeployment, onComplete]);

  const handleKeepWaiting = useCallback(() => {
    startTimeRef.current = Date.now();
    setIsTimedOut(false);
    timeoutExtendedRef.current = true;
  }, []);

  const handleCancel = useCallback(async () => {
    setIsCancelling(true);
    try {
      await deleteDeployment(deployment.id);
    } catch {
      // ignore — deployment may already be gone
    }
    onCancel();
  }, [deployment.id, onCancel]);

  const isFailed = polledDeployment?.status === "failed";

  return (
    <div className="bg-white border border-gray-200 rounded-xl p-6 max-w-md mx-auto">
      <h3 className="text-lg font-semibold mb-4">
        Deploying "{deployment.name}"
      </h3>

      <div className="space-y-0">
        {STEP_LABELS.map((label, i) => (
          <ProgressStep key={label} label={label} status={stepStatuses[i]} />
        ))}
      </div>

      {isFailed && (
        <div className="mt-6 text-center">
          <p className="text-sm text-red-500 mb-3">
            Deployment failed. Check the build command and try again.
          </p>
          <button
            onClick={onFailed}
            className="px-4 py-2 bg-black text-white rounded-lg text-sm hover:bg-gray-800 cursor-pointer"
          >
            Try Again
          </button>
        </div>
      )}

      {isTimedOut && !isFailed && !completedRef.current && (
        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600 mb-3">
            Deployment is taking longer than expected. It may still complete.
          </p>
          <div className="flex justify-center gap-3">
            <button
              onClick={handleKeepWaiting}
              className="px-4 py-2 bg-black text-white rounded-lg text-sm hover:bg-gray-800 cursor-pointer"
            >
              Keep Waiting
            </button>
            <button
              onClick={handleCancel}
              disabled={isCancelling}
              className="px-4 py-2 border border-gray-300 rounded-lg text-sm hover:border-black cursor-pointer disabled:opacity-50"
            >
              {isCancelling ? "Cancelling..." : "Cancel"}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

