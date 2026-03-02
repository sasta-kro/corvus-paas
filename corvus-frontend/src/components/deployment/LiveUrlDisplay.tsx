import { useState, useCallback } from "react";

interface LiveUrlDisplayProps {
  url: string;
}

export default function LiveUrlDisplay({ url }: LiveUrlDisplayProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true); setTimeout(() => setCopied(false), 2000);
    } catch {
      const textarea = document.createElement("textarea");
      textarea.value = url; document.body.appendChild(textarea);
      textarea.select(); document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(true); setTimeout(() => setCopied(false), 2000);
    }
  }, [url]);

  return (
    <div className="flex items-center gap-3 flex-wrap">
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="break-all transition-colors"
        style={{
          color: "var(--sumi)",
          fontSize: "0.95rem",
          textDecoration: "underline",
          textUnderlineOffset: "3px",
          textDecorationColor: "var(--sumi-wash)",
        }}
        onMouseEnter={(e) => (e.currentTarget.style.textDecorationColor = "var(--sumi)")}
        onMouseLeave={(e) => (e.currentTarget.style.textDecorationColor = "var(--sumi-wash)")}
      >
        {url}
      </a>
      <button onClick={handleCopy} className="ink-btn-outline" style={{ padding: "0.25rem 0.6rem", fontSize: "0.8rem" }}>
        {copied ? "Copied!" : "Copy"}
      </button>
    </div>
  );
}
