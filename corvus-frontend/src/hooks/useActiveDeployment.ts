import { useState, useCallback } from "react";
import { STORAGE_KEY_ACTIVE_DEPLOYMENT } from "../config/constants";
import type { ActiveDeploymentSession } from "../types/deployment";

export function useActiveDeployment() {
  const [session, setSession] = useState<ActiveDeploymentSession | null>(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY_ACTIVE_DEPLOYMENT);
      return stored ? (JSON.parse(stored) as ActiveDeploymentSession) : null;
    } catch {
      return null;
    }
  });

  const setActiveDeployment = useCallback((id: string, slug: string) => {
    const value: ActiveDeploymentSession = { id, slug };
    localStorage.setItem(STORAGE_KEY_ACTIVE_DEPLOYMENT, JSON.stringify(value));
    setSession(value);
  }, []);

  const clearActiveDeployment = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY_ACTIVE_DEPLOYMENT);
    setSession(null);
  }, []);

  return {
    activeDeployment: session,
    setActiveDeployment,
    clearActiveDeployment,
    hasActiveDeployment: session !== null,
  };
}

