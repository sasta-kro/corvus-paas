import { useId } from "react";
import InkSpinner from "../shared/InkSpinner";

interface ProgressStepProps {
  label: string;
  status: "completed" | "in-progress" | "pending" | "failed";
}

export default function ProgressStep({ label, status }: ProgressStepProps) {
  const id = useId();
  const gradId = `brush-ink-${id}`;
  const filterId = `brush-tex-${id}`;

  const icon = () => {
    switch (status) {
      case "completed":
        return null;
      case "in-progress":
        return <InkSpinner size="sm" />;
      case "pending":
        return (
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
            {/* Small ink circle — unfilled */}
            <circle cx="9" cy="9" r="3" stroke="var(--sumi-ghost)" strokeWidth="1.5" fill="none" />
          </svg>
        );
      case "failed":
        return (
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
            {/* Brush stroke X */}
            <path d="M5 5L13 13M13 5L5 13" stroke="var(--vermillion)" strokeWidth="2.5" strokeLinecap="round" />
          </svg>
        );
    }
  };

  const textStyle = (): React.CSSProperties => {
    switch (status) {
      case "completed": return { color: "var(--sumi-wash)" };
      case "in-progress": return { color: "var(--sumi)", fontWeight: 700 };
      case "pending": return { color: "var(--sumi-ghost)" };
      case "failed": return { color: "var(--vermillion)", fontWeight: 700 };
    }
  };

  return (
    <div className="flex items-center gap-3 py-2.5" style={{ position: "relative" }}>
      <div className="w-5 h-5 flex items-center justify-center flex-shrink-0">
        {icon()}
      </div>
      <span style={{ fontSize: "0.95rem", ...textStyle() }}>{label}</span>
      {status === "completed" && (
        <svg
          style={{
            position: "absolute",
            left: 0,
            top: "50%",
            width: "100%",
            height: "6px",
            transform: "translateY(-50%)",
            animation: "ink-stroke-draw 0.8s ease-out forwards",
          }}
          viewBox="0 0 100 10"
          preserveAspectRatio="none"
        >
          <defs>
            <linearGradient id={gradId}>
              <stop offset="0%" stopColor="var(--sumi-light)" stopOpacity={0.4} />
              <stop offset="20%" stopColor="var(--sumi-light)" stopOpacity={0.65} />
              <stop offset="50%" stopColor="var(--sumi-light)" stopOpacity={0.8} />
              <stop offset="80%" stopColor="var(--sumi-light)" stopOpacity={0.95} />
              <stop offset="100%" stopColor="var(--sumi-light)" stopOpacity={1.0} />
            </linearGradient>
            <filter id={filterId}>
              <feTurbulence type="fractalNoise" baseFrequency="0.04 0.15" numOctaves="4" seed="3" />
              <feDisplacementMap in="SourceGraphic" scale="2.5" />
            </filter>
          </defs>
          {/* Wedge brush stroke — thin entry, continuously widening, thickest at right edge */}
          <path
            d="M0,5 C3,4 8,3.5 15,3 C30,2.2 50,1.5 70,0.8 C80,0.5 90,0.2 100,0 L100,10 C90,9.8 80,9.5 70,9.2 C50,8.5 30,7.8 15,7 C8,6.5 3,6 0,5Z"
            fill={`url(#${gradId})`}
            filter={`url(#${filterId})`}
          />
        </svg>
      )}
    </div>
  );
}
