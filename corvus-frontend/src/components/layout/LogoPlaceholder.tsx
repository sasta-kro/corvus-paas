/** Corvus logo with official SVG asset */
export default function LogoPlaceholder() {
  return (
    <div className="flex items-center gap-3">
      <img
        src="/corvus-logo1-optimized.svg"
        alt="Corvus Logo"
        width={32}
        // height={35}
        style={
          {
              display: "block",
            // outline: "cyan solid 1px" // for debugging
          }
      }
      />

      {/* Name */}
      <div className="relative">
        <span
          className="font-brush text-2xl"
          style={{
            color: "var(--sumi)",
            textShadow: "1px 1px 0 rgba(17,17,16,0.08)",
          }}
        >
          CorvusPaaS
        </span>
      </div>
    </div>
  );
}
