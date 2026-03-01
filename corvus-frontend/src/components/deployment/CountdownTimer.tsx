import { useEffect, useRef } from "react";
import { useCountdown } from "../../hooks/useCountdown";

interface CountdownTimerProps {
  expiresAt: Date;
  onExpired: () => void;
}

/** Countdown timer showing MM:SS until deployment expires */
export default function CountdownTimer({
  expiresAt,
  onExpired,
}: CountdownTimerProps) {
  const { formattedTime, isExpired, timeRemaining } = useCountdown(expiresAt);
  const expiredCalledRef = useRef(false);

  // Call onExpired via useEffect to avoid calling during render
  useEffect(() => {
    if (isExpired && !expiredCalledRef.current) {
      expiredCalledRef.current = true;
      onExpired();
    }
  }, [isExpired, onExpired]);

  if (isExpired) {
    return <span className="text-sm text-red-500 font-medium">Expired</span>;
  }

  const isWarning = timeRemaining && timeRemaining.minutes < 2;

  return (
    <span
      className={`text-sm font-mono ${
        isWarning ? "text-red-500 font-medium" : "text-gray-600"
      }`}
    >
      Expires in {formattedTime}
    </span>
  );
}

