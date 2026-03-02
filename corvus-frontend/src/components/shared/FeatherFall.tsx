/** Ambient falling crow feathers using actual ink-drawn feather SVG assets */

const FEATHER_ASSETS = [
  "/fluffy-feather-optimized.svg",
  "/long-feather-optimized.svg",
  "/round-feather-optimized.svg",
];

const ANIMATION_NAMES = ["feather-drift", "feather-drift-2", "feather-drift-3"];

interface FeatherData {
  id: number;
  asset: number;
  left: number;
  delay: number;
  duration: number;
  size: number;
  animIndex: number;
  opacity: number;
  rotation: number;
  flipX: boolean;
}

const feathers: FeatherData[] = Array.from({ length: 30 }, (_, i) => ({
  id: i,
  asset: i % FEATHER_ASSETS.length,
  left: 1 + ((i * 3.3 + 2) % 95),
  delay: i * 1.8 + Math.sin(i * 1.4) * 3,
  duration: 20 + (i % 7) * 5,
  size: 32 + (i % 6) * 14,
  animIndex: i % 3,
  opacity: 0.4 + (i % 5) * 0.06,
  rotation: ((i * 47 + 13) % 360) - 180,
  flipX: i % 2 === 0,
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
          <img
            src={FEATHER_ASSETS[f.asset]}
            alt=""
            style={{
              width: "100%",
              height: "100%",
              objectFit: "contain",
              transform: `rotate(${f.rotation}deg)${f.flipX ? " scaleX(-1)" : ""}`,
              animation: `${ANIMATION_NAMES[f.animIndex]} ${f.duration}s ${f.delay}s linear infinite backwards`,
              filter: "brightness(0) opacity(0.7)",
            }}
          />
        </div>
      ))}
    </div>
  );
}
