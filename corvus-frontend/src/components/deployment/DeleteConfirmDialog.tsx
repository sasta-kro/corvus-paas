import * as Dialog from "@radix-ui/react-dialog";
import InkSpinner from "../shared/InkSpinner";

interface DeleteConfirmDialogProps {
  isOpen: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  isDeleting: boolean;
}

export default function DeleteConfirmDialog({ isOpen, onConfirm, onCancel, isDeleting }: DeleteConfirmDialogProps) {
  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => !open && onCancel()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40" style={{ backgroundColor: "rgba(17,17,16,0.55)" }} />
        <Dialog.Content
          style={{
            position: "fixed",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            width: "100%",
            maxWidth: "24rem",
            zIndex: 50,
          }}
        >
          <div className="ink-card torn-edge-3">
            <Dialog.Title className="font-brush text-lg mb-2" style={{ color: "var(--sumi)" }}>
              Delete Deployment
            </Dialog.Title>
            <Dialog.Description style={{ color: "var(--sumi-light)", fontSize: "0.9rem", marginBottom: "1.5rem", fontStyle: "italic" }}>
              Are you sure? This cannot be undone.
            </Dialog.Description>

            <div className="flex justify-end gap-3">
              <button onClick={onCancel} disabled={isDeleting} className="ink-btn-outline">Cancel</button>
              <button onClick={onConfirm} disabled={isDeleting} className="ink-btn-danger flex items-center gap-2">
                {isDeleting && <InkSpinner size="sm" />}
                Delete
              </button>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
