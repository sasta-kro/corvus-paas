interface DeployButtonProps {
  onClick: () => void;
  disabled: boolean;
  loading: boolean;
  label?: string;
}

/** Shared deploy trigger button */
export default function DeployButton({
  onClick,
  disabled,
  loading,
  label = "Deploy",
}: DeployButtonProps) {
  return (
    <div className="flex justify-end">
      <button
        onClick={onClick}
        disabled={disabled}
        className="px-6 py-2.5 bg-black text-white rounded-lg text-sm font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer transition-colors"
      >
        {loading ? (
          <span className="flex items-center gap-2">
            <svg
              className="animate-spin h-4 w-4"
              viewBox="0 0 24 24"
              fill="none"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            Deploying...
          </span>
        ) : (
          label
        )}
      </button>
    </div>
  );
}

