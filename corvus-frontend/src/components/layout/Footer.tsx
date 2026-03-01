/** Footer with links and tech credits */
export default function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-white mt-auto">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-6 text-center">
        <div className="flex flex-wrap justify-center gap-4 text-sm text-gray-500 mb-3">
          <a
            href="https://github.com/sasta-kro/corvus-paas"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-black transition-colors"
          >
            GitHub
          </a>
          <a
            href="https://github.com/sasta-kro"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-black transition-colors"
          >
            @sasta-kro
          </a>
          <a
            href="https://linkedin.com"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-black transition-colors"
          >
            LinkedIn
          </a>
        </div>
        <p className="text-xs text-gray-400">
          Built with Go, Docker, and Traefik
        </p>
      </div>
    </footer>
  );
}

