import InkSpinner from "../shared/InkSpinner";

interface DeployButtonProps {
  onClick: () => void;
  disabled: boolean;
  loading: boolean;
  label?: string;
}

export default function DeployButton({ onClick, disabled, loading, label = "Deploy" }: DeployButtonProps) {
  return (
    <div className="flex justify-end">
      <button onClick={onClick} disabled={disabled} className="ink-btn">
        {loading ? (
          <span className="flex items-center gap-2">
            <InkSpinner size="sm" />
            Deploying...
          </span>
        ) : (
          label
        )}
      </button>
    </div>
  );
}
