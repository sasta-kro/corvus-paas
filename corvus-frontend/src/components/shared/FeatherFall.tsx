/** Ambient falling crow feathers — realistic barbed feather shapes */

/* Detailed crow feather SVG paths — each slightly different */
const FEATHER_SHAPES = [
  // Long flight feather — tapered with barb lines
  `M12 1C12 1 10.5 4 10 8C9.5 12 9.8 16 10.5 20C10.8 22 11.5 23.5 12 24
   C12.5 23.5 13.2 22 13.5 20C14.2 16 14.5 12 14 8C13.5 4 12 1 12 1Z
   M12 4L10.2 9 M12 7L10 12 M12 10L9.8 15 M12 13L10.2 17 M12 16L10.5 19
   M12 4L13.8 9 M12 7L14 12 M12 10L14.2 15 M12 13L13.8 17 M12 16L13.5 19`,
  // Shorter body feather — rounder
  `M12 3C12 3 9 7 8.5 11C8 15 9 19 10 21C10.8 22.5 11.5 23 12 23
   C12.5 23 13.2 22.5 14 21C15 19 16 15 15.5 11C15 7 12 3 12 3Z
   M12 6L9.5 10 M12 9L9 13 M12 12L9.5 16 M12 15L10 18
   M12 6L14.5 10 M12 9L15 13 M12 12L14.5 16 M12 15L14 18`,
  // Small downy feather — fluffier
  `M12 5C12 5 9.5 8 9 11C8.5 14 9.5 17 10.5 19.5C11 20.5 11.6 21 12 21
   C12.4 21 13 20.5 13.5 19.5C14.5 17 15.5 14 15 11C14.5 8 12 5 12 5Z
   M12 7L10 10 M12 9.5L9.5 13 M12 12L10 15 M12 14.5L10.5 17
   M12 7L14 10 M12 9.5L14.5 13 M12 12L14 15 M12 14.5L13.5 17`,
];

const ANIMATION_NAMES = ["feather-drift", "feather-drift-2", "feather-drift-3"];

interface FeatherData {
  id: number;
  shape: number;
  left: number;
  delay: number;
  duration: number;
  size: number;
  animIndex: number;
  opacity: number;
}

const feathers: FeatherData[] = Array.from({ length: 18 }, (_, i) => ({
  id: i,
  shape: i % FEATHER_SHAPES.length,
  left: 2 + ((i * 5.3 + 3) % 92),
  delay: i * 2.7 + Math.sin(i * 1.4) * 4,
  duration: 28 + (i % 7) * 6,
  size: 24 + (i % 6) * 10,
  animIndex: i % 3,
  opacity: 0.12 + (i % 5) * 0.05,
}));

export default function FeatherFall() {
  return (
    <div
      className="fixed inset-0 pointer-events-none overflow-hidden"
      style={{ zIndex: 0 }}
    >
      {feathers.map((f) => (
        <div
          key={f.id}
          style={{
            position: "absolute",
            left: `${f.left}%`,
            top: 0,
            width: f.size,
            height: f.size * 1.5,
            opacity: f.opacity,
          }}
        >
          <svg
            style={{
              width: "100%",
              height: "100%",
              animation: `${ANIMATION_NAMES[f.animIndex]} ${f.duration}s ${f.delay}s linear infinite backwards`,
            }}
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--sumi)"
            strokeWidth={f.size > 50 ? 0.5 : 0.8}
          >
            <path d={FEATHER_SHAPES[f.shape]} fill="var(--sumi)" fillOpacity="0.7" />
          </svg>
        </div>
      ))}
    </div>
  );
}
