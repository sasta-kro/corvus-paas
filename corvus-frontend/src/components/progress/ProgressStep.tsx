interface ProgressStepProps {
  label: string;
  status: "completed" | "in-progress" | "pending" | "failed";
}

/** Single progress step row with status icon and label */
export default function ProgressStep({ label, status }: ProgressStepProps) {
  const icon = () => {
    switch (status) {
      case "completed":
        return <span className="text-gray-400">✓</span>;
      case "in-progress":
        return (
          <svg
            className="animate-spin h-4 w-4 text-black"
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
        );
      case "pending":
        return <span className="text-gray-300">○</span>;
      case "failed":
        return <span className="text-red-500">✗</span>;
    }
  };

  const textClass = () => {
    switch (status) {
      case "completed":
        return "text-gray-400";
      case "in-progress":
        return "text-black font-medium";
      case "pending":
        return "text-gray-300";
      case "failed":
        return "text-red-500";
    }
  };

  return (
    <div className="flex items-center gap-3 py-2">
      <div className="w-5 h-5 flex items-center justify-center text-sm">
        {icon()}
      </div>
      <span className={`text-sm ${textClass()}`}>{label}</span>
    </div>
  );
}

