import { useState, useEffect, useMemo } from "react";
import { formatCountdown } from "../lib/utils";

interface UseCountdownResult {
  timeRemaining: { minutes: number; seconds: number } | null;
  isExpired: boolean;
  formattedTime: string;
}

export function useCountdown(expiresAt: Date | null): UseCountdownResult {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!expiresAt) return;

    const interval = setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => clearInterval(interval);
  }, [expiresAt]);

  return useMemo(() => {
    if (!expiresAt) {
      return { timeRemaining: null, isExpired: false, formattedTime: "--:--" };
    }

    const diff = expiresAt.getTime() - now;

    if (diff <= 0) {
      return { timeRemaining: null, isExpired: true, formattedTime: "00:00" };
    }

    const totalSeconds = Math.floor(diff / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;

    return {
      timeRemaining: { minutes, seconds },
      isExpired: false,
      formattedTime: formatCountdown(minutes, seconds),
    };
  }, [expiresAt, now]);
}

