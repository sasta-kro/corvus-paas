import { useState, useEffect, useRef, useCallback } from "react";
import { getDeployment } from "../api/deployments";
import { ApiError } from "../api/client";
import { POLL_INTERVAL_MS } from "../config/constants";
import type { Deployment } from "../types/deployment";

interface UseDeploymentPollingResult {
  deployment: Deployment | null;
  isLoading: boolean;
  error: ApiError | null;
  isNotFound: boolean;
  refetch: () => void;
}

export function useDeploymentPolling(
  deploymentId: string | null,
  intervalMs: number = POLL_INTERVAL_MS,
  enabled: boolean = true
): UseDeploymentPollingResult {
  const [deployment, setDeployment] = useState<Deployment | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<ApiError | null>(null);
  const [isNotFound, setIsNotFound] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const stoppedRef = useRef(false);

  const fetchDeployment = useCallback(async () => {
    if (!deploymentId || stoppedRef.current) return;
    try {
      const data = await getDeployment(deploymentId);
      setDeployment(data);
      setError(null);
      setIsNotFound(false);
      setIsLoading(false);
      // Stop polling on terminal states
      if (data.status === "live" || data.status === "failed") {
        stoppedRef.current = true;
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      }
    } catch (err) {
      setIsLoading(false);
      if (err instanceof ApiError && err.status === 404) {
        setIsNotFound(true);
        stoppedRef.current = true;
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
      } else {
        setError(err instanceof ApiError ? err : new ApiError("Network error", 0));
      }
    }
  }, [deploymentId]);

  useEffect(() => {
    stoppedRef.current = false;
    setDeployment(null);
    setIsLoading(true);
    setError(null);
    setIsNotFound(false);

    if (!deploymentId || !enabled) {
      setIsLoading(false);
      return;
    }

    // Immediate fetch
    fetchDeployment();

    // Start polling
    intervalRef.current = setInterval(fetchDeployment, intervalMs);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [deploymentId, intervalMs, enabled, fetchDeployment]);

  return { deployment, isLoading, error, isNotFound, refetch: fetchDeployment };
}

