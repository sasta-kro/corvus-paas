import InkSplatter from "../shared/InkSplatter";

export default function HeroSection() {
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
        A self-hosted PaaS platform. Pick a preset, upload a zip,
        or paste a GitHub URL.
      </p>

      {/* Expressive dual-stroke brush divider */}
      <div className="mt-8 mx-auto" style={{ maxWidth: 260 }}>
        <svg width="100%" height="12" viewBox="0 0 260 12" preserveAspectRatio="none" fill="none">
          {/* Primary S-curved organic stroke */}
          <path
            d="M0,6 C30,2 60,10 90,5 C120,0 150,11 180,6 C210,1 240,9 260,5"
            stroke="var(--sumi)"
            strokeWidth={3}
            strokeLinecap="round"
            opacity={0.5}
            fill="none"
          />
          {/* Secondary ghost stroke — slightly offset */}
          <path
            d="M0,7 C35,3 65,11 95,6 C125,1 155,10 185,5 C215,2 245,8 260,6"
            stroke="var(--sumi)"
            strokeWidth={1}
            strokeLinecap="round"
            opacity={0.2}
            fill="none"
          />
        </svg>
      </div>
    </section>
  );
}
