import { useState, useCallback } from "react";

interface LiveUrlDisplayProps {
  url: string;
}

/** Displays the deployment URL with a copy button */
export default function LiveUrlDisplay({ url }: LiveUrlDisplayProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const textarea = document.createElement("textarea");
      textarea.value = url;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [url]);

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="text-sm text-black underline hover:text-gray-600 break-all"
      >
        {url}
      </a>
      <button
        onClick={handleCopy}
        className="px-2 py-1 text-xs border border-gray-300 rounded hover:border-black transition-colors cursor-pointer"
      >
        {copied ? "Copied!" : "Copy"}
      </button>
    </div>
  );
}

