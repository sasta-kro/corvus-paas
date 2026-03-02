/** Ink brush-circle loading spinner — like an ensō being drawn */

interface InkSpinnerProps {
  size?: "sm" | "md" | "lg";
  className?: string;
}

const dims = { sm: 20, md: 36, lg: 52 };

export default function InkSpinner({ size = "md", className = "" }: InkSpinnerProps) {
  const d = dims[size];
  const cx = d / 2;
  const cy = d / 2;
  const r = d / 2 - 4;
  const circumference = 2 * Math.PI * r;

  return (
    <svg
      width={d}
      height={d}
      viewBox={`0 0 ${d} ${d}`}
      className={className}
      style={{ display: "inline-block" }}
    >
      {/* Faint full circle — ghost ring */}
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke="var(--sumi-ghost)"
        strokeWidth={size === "sm" ? 1.5 : 2}
        opacity="0.3"
      />
      {/* Animated brush stroke circle — ensō */}
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke="var(--sumi)"
        strokeWidth={size === "sm" ? 2 : 3}
        strokeLinecap="round"
        strokeDasharray={circumference}
        style={{
          animation: "ink-circle-draw 55s cubic-bezier(0.35, 0, 0.25, 1) infinite",
          transformOrigin: "center",
        }}
      />
      {/* Center ink dot — breathing */}
      <circle
        cx={cx}
        cy={cy}
        r={size === "sm" ? 1.5 : 3}
        fill="var(--sumi)"
        style={{
          animation: "ink-dot-breathe 55s cubic-bezier(0.35, 0, 0.25, 1) infinite",
        }}
      />
    </svg>
  );
}
