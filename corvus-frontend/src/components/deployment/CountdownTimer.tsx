import { useEffect, useRef } from "react";
import { useCountdown } from "../../hooks/useCountdown";

interface CountdownTimerProps {
  expiresAt: Date;
  onExpired: () => void;
}

export default function CountdownTimer({ expiresAt, onExpired }: CountdownTimerProps) {
  const { formattedTime, isExpired, timeRemaining } = useCountdown(expiresAt);
  const expiredCalledRef = useRef(false);

  useEffect(() => {
    if (isExpired && !expiredCalledRef.current) { expiredCalledRef.current = true; onExpired(); }
  }, [isExpired, onExpired]);

  if (isExpired) {
    return <span style={{ color: "var(--vermillion)", fontWeight: 700, fontSize: "0.9rem" }}>Expired</span>;
  }

  const isWarning = timeRemaining && timeRemaining.minutes < 2;

  return (
    <span style={{
      fontFamily: "monospace",
      fontSize: "0.9rem",
      fontWeight: isWarning ? 700 : 400,
      color: isWarning ? "var(--vermillion)" : "var(--sumi-light)",
    }}>
      Expires in {formattedTime}
    </span>
  );
}
