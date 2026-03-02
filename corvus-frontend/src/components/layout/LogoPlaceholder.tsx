/** Brushstroke crow logo with vermillion seal */
export default function LogoPlaceholder() {
  return (
    <div className="flex items-center gap-3">
      {/* Sumi-e crow — bold brushstroke silhouette */}
      <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
        {/* Body — thick ink stroke */}
        <path
          d="M8 20C8 14 12 8 16 6C20 8 24 14 24 20C24 23 22 25 20 26L18 22L16 27L14 22L12 26C10 25 8 23 8 20Z"
          fill="var(--sumi)"
          stroke="var(--sumi)"
          strokeWidth="0.5"
        />
        {/* Head */}
        <circle cx="13" cy="13" r="4" fill="var(--sumi)" />
        {/* Eye — paper color gleam */}
        <circle cx="12" cy="12.5" r="1" fill="var(--paper)" />
        <circle cx="12.2" cy="12.3" r="0.3" fill="var(--sumi)" />
        {/* Beak — sharp ink stroke */}
        <path d="M9 13L6 14.5L9 14Z" fill="var(--sumi-mid)" />
        {/* Wing texture — brush lines */}
        <path d="M17 15Q20 13 24 15" stroke="var(--paper-aged)" strokeWidth="0.6" fill="none" />
        <path d="M16 17Q19 15.5 23 17.5" stroke="var(--paper-aged)" strokeWidth="0.4" fill="none" />
        {/* Tail feathers */}
        <path d="M14 26L13 30M16 27L16 31M18 26L19 30" stroke="var(--sumi)" strokeWidth="1" strokeLinecap="round" />
      </svg>

      {/* Name + seal */}
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
        {/* Vermillion hanko seal (chinese red seal)*/}
        {/*<svg*/}
        {/*  width="14"*/}
        {/*  height="14"*/}
        {/*  viewBox="0 0 14 14"*/}
        {/*  className="absolute -top-1 -right-4"*/}
        {/*  style={{ opacity: 0.85 }}*/}
        {/*>*/}
        {/*  <rect x="1" y="1" width="12" height="12" rx="1" fill="var(--vermillion)" />*/}
        {/*  <text*/}
        {/*    x="7"*/}
        {/*    y="10"*/}
        {/*    textAnchor="middle"*/}
        {/*    fill="var(--paper)"*/}
        {/*    fontSize="7"*/}
        {/*    fontWeight="bold"*/}
        {/*    fontFamily="serif"*/}
        {/*  >*/}
        {/*    鴉*/}
        {/*  </text>*/}
        {/*</svg>*/}
      </div>
    </div>
  );
}
