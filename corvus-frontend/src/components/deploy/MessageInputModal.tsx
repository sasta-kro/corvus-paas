import { useState, useRef, useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { MAX_MESSAGE_LENGTH } from "../../config/constants";

interface MessageInputModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (message: string) => void;
}

export default function MessageInputModal({ isOpen, onClose, onSubmit }: MessageInputModalProps) {
  const [message, setMessage] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) { setMessage(""); setTimeout(() => inputRef.current?.focus(), 100); }
  }, [isOpen]);

  const handleSubmit = () => { if (message.trim()) { onSubmit(message.trim()); setMessage(""); } };

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
            maxWidth: "28rem",
            zIndex: 50,
          }}
          className="ink-card torn-edge-2"
        >
          <Dialog.Title className="font-brush text-lg mb-1" style={{ color: "var(--sumi)" }}>
            Your Message
          </Dialog.Title>
          <Dialog.Description style={{ color: "var(--sumi-light)", fontSize: "0.9rem", marginBottom: "1rem", fontStyle: "italic" }}>
            What should your page say?
          </Dialog.Description>

          <input ref={inputRef} type="text" value={message}
            onChange={(e) => setMessage(e.target.value.slice(0, MAX_MESSAGE_LENGTH))}
            onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
            placeholder="Enter your message..." maxLength={MAX_MESSAGE_LENGTH} className="ink-input" />
          <div style={{ textAlign: "right", color: "var(--sumi-ghost)", fontSize: "0.75rem", marginTop: "0.3rem" }}>
            {message.length} / {MAX_MESSAGE_LENGTH}
          </div>

          <div className="flex justify-end gap-3 mt-5">
            <button onClick={onClose} className="ink-btn-outline">Cancel</button>
            <button onClick={handleSubmit} disabled={!message.trim()} className="ink-btn">Deploy My Message</button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
