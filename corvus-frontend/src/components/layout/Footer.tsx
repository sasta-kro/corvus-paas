import InkSplatter from "../shared/InkSplatter";

export default function Footer() {
  return (
    <footer className="mt-auto relative" style={{ zIndex: 10 }}>
      <div className="brush-divider-h" />
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8 text-center relative">
        <InkSplatter variant={1} size={45} style={{ bottom: 10, right: "8%", opacity: 0.08 }} />

        <div className="flex flex-wrap justify-center gap-6 text-sm mb-4">
          {[
            { href: "https://github.com/sasta-kro/corvus-paas", label: "GitHub" },
            { href: "https://github.com/sasta-kro", label: "@sasta-kro" },
            { href: "https://linkedin.com", label: "LinkedIn" },
          ].map((link) => (
            <a
              key={link.label}
              href={link.href}
              target="_blank"
              rel="noopener noreferrer"
              className="transition-colors"
              style={{ color: "var(--sumi-wash)", fontWeight: 700 }}
              onMouseEnter={(e) => (e.currentTarget.style.color = "var(--sumi)")}
              onMouseLeave={(e) => (e.currentTarget.style.color = "var(--sumi-wash)")}
            >
              {link.label}
            </a>
          ))}
        </div>
        <p style={{ color: "var(--sumi-ghost)", fontSize: "0.8rem", fontStyle: "italic" }}>
          Built with Go, Docker, and Traefik
        </p>
      </div>
    </footer>
  );
}
