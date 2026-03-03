import * as Dialog from "@radix-ui/react-dialog";

interface BuildImageWarningModalProps {
  isOpen: boolean;
  onClose: () => void;
  onDeployAnyway: () => void;
}

/**
 * Shown when the user types a build command that doesn't look like a Node.js
 * toolchain command. The current build environment is node:20-alpine, so only
 * npm / yarn / pnpm / npx / node commands are natively supported for now.
 */
export default function BuildImageWarningModal({ isOpen, onClose, onDeployAnyway }: BuildImageWarningModalProps) {
  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Portal container={document.body}>
        <Dialog.Overlay className="fixed inset-0 z-40" style={{ backgroundColor: "rgba(17,17,16,0.55)" }} />
        <Dialog.Content
          style={{
            position: "fixed",
            top: "50%",
            left: "50%",
            transform: "translate(-50%, -50%)",
            width: "100%",
            maxWidth: "30rem",
            zIndex: 50,
          }}
          className="ink-card torn-edge-2"
        >
          <Dialog.Title className="font-brush text-lg mb-1" style={{ color: "var(--sumi)" }}>
            Heads up
          </Dialog.Title>
          <Dialog.Description
            style={{ color: "var(--sumi-light)", fontSize: "0.9rem", marginBottom: "1rem", fontStyle: "italic" }}
          >
            That build command might not work out of the box.
          </Dialog.Description>

          <p style={{ color: "var(--sumi)", fontSize: "0.9rem", lineHeight: 1.6, marginBottom: "0.75rem" }}>
            Right now, Corvus builds everything inside a <strong>node:20-alpine</strong> container, so the build
            environment natively understands <code style={{ fontFamily: "monospace", fontSize: "0.85em" }}>npm</code>,{" "}
            <code style={{ fontFamily: "monospace", fontSize: "0.85em" }}>yarn</code>,{" "}
            <code style={{ fontFamily: "monospace", fontSize: "0.85em" }}>pnpm</code>, and other Node.js toolchain commands.
          </p>
          <p style={{ color: "var(--sumi-light)", fontSize: "0.85rem", lineHeight: 1.6, fontStyle: "italic" }}>
            Support for custom build images is on the roadmap. The developer is aware and genuinely wants to add it
            — just has some rather important finals to get through first. It will land soon.
          </p>

          <div className="flex justify-end gap-3 mt-5">
            <button onClick={onClose} className="ink-btn-outline">
              Go back
            </button>
            <button onClick={() => { onDeployAnyway(); onClose(); }} className="ink-btn">
              Deploy anyway
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

