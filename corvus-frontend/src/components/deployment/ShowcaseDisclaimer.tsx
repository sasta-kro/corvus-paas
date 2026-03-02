import { useState, useRef, useEffect } from "react";

/**
 * Info icon with a speech bubble that explains the expiry timer.
 * Appears on hover (desktop) or click (mobile).
 */
export default function ShowcaseDisclaimer() {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  // Close when clicking outside
  useEffect(() => {
    if (!isOpen) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [isOpen]);

  const handleMouseEnter = () => {
    clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => setIsOpen(true), 300);
  };

  const handleMouseLeave = () => {
    clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => setIsOpen(false), 200);
  };

  return (
    <div
      ref={containerRef}
      className="inline-flex items-center"
      style={{ position: "relative", verticalAlign: "middle" }}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {/* Info icon — small ink circle with "?" */}
      <button
        onClick={() => setIsOpen((v) => !v)}
        aria-label="Why is there an expiry timer?"
        className="cursor-pointer"
        style={{
          background: "none",
          border: "none",
          padding: "2px",
          display: "inline-flex",
          alignItems: "center",
          color: "var(--sumi-ghost)",
          transition: "color 0.2s",
        }}
        onMouseEnter={(e) => (e.currentTarget.style.color = "var(--sumi-light)")}
        onMouseLeave={(e) => (e.currentTarget.style.color = "var(--sumi-ghost)")}
      >
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="7" cy="7" r="6" stroke="currentColor" strokeWidth="1.5" />
          <text
            x="7"
            y="10.5"
            textAnchor="middle"
            fill="currentColor"
            fontSize="8.5"
            fontFamily="'EB Garamond', serif"
            fontWeight="700"
            fontStyle="italic"
          >
            ?
          </text>
        </svg>
      </button>

      {/* Speech bubble */}
      {isOpen && (
        <div
          style={{
            position: "absolute",
            bottom: "calc(100% + 10px)",
            left: "50%",
            transform: "translateX(-50%)",
            width: "280px",
            padding: "0.85rem 1rem",
            background: "var(--paper-warm)",
            border: "1.5px solid var(--sumi-ghost)",
            borderRadius: "2px",
            boxShadow: "3px 4px 0 -1px rgba(17,17,16,0.12)",
            zIndex: 50,
            animation: "toast-brush-in 0.25s ease-out",
          }}
        >
          {/* Speech bubble tail */}
          <div
            style={{
              position: "absolute",
              bottom: "-7px",
              left: "50%",
              transform: "translateX(-50%) rotate(45deg)",
              width: "12px",
              height: "12px",
              background: "var(--paper-warm)",
              borderRight: "1.5px solid var(--sumi-ghost)",
              borderBottom: "1.5px solid var(--sumi-ghost)",
            }}
          />
          <p
            style={{
              color: "var(--sumi-light)",
              fontSize: "0.8rem",
              lineHeight: "1.5",
              margin: 0,
              fontStyle: "italic",
            }}
          >
            This is a showcase environment with limited hardware resources. These instances
            are hosted on a personal homelab, which is why they have a short lifespan.
          </p>
          <p
            style={{
              color: "var(--sumi-wash)",
              fontSize: "0.75rem",
              lineHeight: "1.4",
              margin: "0.5rem 0 0 0",
            }}
          >
            The software itself can handle production-grade PaaS workloads. The only limit is the hardware you give it.
          </p>
        </div>
      )}
    </div>
  );
}

