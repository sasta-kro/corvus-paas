import InkSplatter from "../shared/InkSplatter";
import {useId} from "react";

export default function HeroSection() {

    const id = useId();
    const gradId = `hero-brush-${id}`;
    const filterId = `hero-tex-${id}`;

  return (
    <section className="text-center py-14 px-4 sm:px-6 relative" style={{ zIndex: 10 }}>
      {/* Decorative ink splatters */}
      <InkSplatter variant={0} size={70} style={{ top: 20, left: "10%", opacity: 0.08 }} />
      <InkSplatter variant={2} size={55} style={{ top: 60, right: "12%", opacity: 0.07 }} />
      <InkSplatter variant={1} size={45} style={{ bottom: 30, left: "5%", opacity: 0.06 }} />

      {/* Ensō — large faded ink circle behind heading */}
      <svg
        width="320"
        height="320"
        viewBox="0 0 320 320"
        fill="none"
        style={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          opacity: 0.045,
          pointerEvents: "none",
        }}
      >
        <circle
          cx="160"
          cy="160"
          r="140"
          stroke="var(--sumi)"
          strokeWidth={18}
          strokeDasharray="780 100"
          strokeLinecap="round"
          fill="none"
        />
      </svg>

      <h1
        className="font-brush text-5xl sm:text-6xl mb-5 leading-tight"
        style={{
          color: "var(--sumi)",
          textShadow: "1px 1px 0 rgba(17,17,16,0.15), 0 0 8px rgba(17,17,16,0.04)",
        }}
      >
        Deploy a website
        <br />
        in seconds.
      </h1>
      <p
        className="text-lg sm:text-xl max-w-lg mx-auto leading-relaxed"
        style={{
          color: "var(--sumi-light)",
          fontStyle: "italic",
        }}
      >
        A self-hosted PaaS platform. Pick a preset to try out,
          upload a zip of your build, or paste a GitHub URL.
      </p>

      {/* Hand-drawn ink brush stroke divider */}
      <div className="mt-8 mx-auto" style={{ maxWidth: 260 }}>
        <svg width="100%" height="14" viewBox="0 0 260 14" preserveAspectRatio="none" fill="none">
          <defs>
            <linearGradient id={gradId}>
              <stop offset="0%" stopColor="var(--sumi)" stopOpacity={0} />
              <stop offset="8%" stopColor="var(--sumi)" stopOpacity={0.6} />
              <stop offset="20%" stopColor="var(--sumi)" stopOpacity={0.85} />
              <stop offset="50%" stopColor="var(--sumi)" stopOpacity={0.9} />
              <stop offset="80%" stopColor="var(--sumi)" stopOpacity={0.7} />
              <stop offset="95%" stopColor="var(--sumi)" stopOpacity={0.3} />
              <stop offset="100%" stopColor="var(--sumi)" stopOpacity={0} />
            </linearGradient>
            <filter id={filterId}>
              <feTurbulence type="fractalNoise" baseFrequency="0.04 0.12" numOctaves="4" seed="7" />
              <feDisplacementMap in="SourceGraphic" scale="3" />
            </filter>
          </defs>
          {/* Brush flick stroke — tapers on both ends, thickest in center, with organic distortion */}
          <path
            d="M0,7 C8,5 20,3.5 40,3 C70,2 100,1.5 130,1 C160,1.5 190,2.5 220,4 C240,5 252,6 260,7
               C252,8 240,9.5 220,10.5 C190,11.5 160,12.5 130,13 C100,12.5 70,11.5 40,10.5 C20,9.5 8,8.5 0,7Z"
            fill={`url(#${gradId})`}
            filter={`url(#${filterId})`}
          />
        </svg>
      </div>
    </section>
  );
}
