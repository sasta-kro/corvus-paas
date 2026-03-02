/** Decorative ink splatter marks — scattered on page for painterly feel */

const SPLATTERS = [
  // Small splat cluster
  { cx: 10, cy: 10, paths: "M8 10a2 2 0 1 0 4 0 2 2 0 1 0-4 0M6 8a1 1 0 1 0 2 0 1 1 0 1 0-2 0M13 7a0.7 0.7 0 1 0 1.4 0 0.7 0.7 0 1 0-1.4 0M5 12a0.5 0.5 0 1 0 1 0 0.5 0.5 0 1 0-1 0" },
  // Drip streak
  { cx: 10, cy: 12, paths: "M9 4a1.5 1.5 0 1 0 3 0 1.5 1.5 0 1 0-3 0M10 6L9.5 14Q9 18 10 20Q11 18 10.5 14Z" },
  // Dot scatter
  { cx: 10, cy: 10, paths: "M4 6a1 1 0 1 0 2 0 1 1 0 1 0-2 0M8 3a1.3 1.3 0 1 0 2.6 0 1.3 1.3 0 1 0-2.6 0M14 5a0.8 0.8 0 1 0 1.6 0 0.8 0.8 0 1 0-1.6 0M6 14a0.6 0.6 0 1 0 1.2 0 0.6 0.6 0 1 0-1.2 0M12 12a1.5 1.5 0 1 0 3 0 1.5 1.5 0 1 0-3 0M15 15a0.4 0.4 0 1 0 0.8 0 0.4 0.4 0 1 0-0.8 0" },
];

interface InkSplatterProps {
  variant?: 0 | 1 | 2;
  size?: number;
  className?: string;
  style?: React.CSSProperties;
}

export default function InkSplatter({
  variant = 0,
  size = 40,
  className = "",
  style = {},
}: InkSplatterProps) {
  const s = SPLATTERS[variant];
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 20 20"
      className={`ink-splatter ${className}`}
      style={style}
      fill="var(--sumi)"
    >
      <path d={s.paths} />
    </svg>
  );
}
