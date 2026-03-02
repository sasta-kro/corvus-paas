import { createContext, useContext, useState, useCallback, type ReactNode } from "react";

type ToastType = "success" | "error" | "info";

interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

interface ToastContextType {
  toasts: Toast[];
  addToast: (message: string, type?: ToastType) => void;
  removeToast: (id: number) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

let toastId = 0;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback((message: string, type: ToastType = "info") => {
    const id = ++toastId;
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => removeToast(id), 4000);
  }, [removeToast]);

  const toastStyle = (type: ToastType): React.CSSProperties => {
    switch (type) {
      case "success": return {
        background: "var(--leaf-bg)", border: "2px solid var(--leaf)", color: "var(--leaf)",
      };
      case "error": return {
        background: "var(--vermillion-bg)", border: "2px solid var(--vermillion)", color: "var(--vermillion)",
      };
      default: return {
        background: "var(--paper)", border: "2px solid var(--sumi-light)", color: "var(--sumi)",
      };
    }
  };

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
      <div className="fixed top-4 right-4 z-50 flex flex-col gap-2" style={{ maxWidth: "22rem" }}>
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className="px-4 py-3 flex items-center justify-between gap-3 animate-toast-in"
            style={{
              ...toastStyle(toast.type),
              borderRadius: "1px 3px 2px 1px",
              fontFamily: '"EB Garamond", serif',
              fontSize: "0.9rem",
              fontWeight: 700,
              boxShadow: "3px 3px 0 rgba(17,17,16,0.08)",
            }}
          >
            <span>{toast.message}</span>
            <button
              onClick={() => removeToast(toast.id)}
              className="cursor-pointer"
              style={{ lineHeight: 1, opacity: 0.6 }}
              onMouseEnter={(e) => (e.currentTarget.style.opacity = "1")}
              onMouseLeave={(e) => (e.currentTarget.style.opacity = "0.6")}
            >
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M3 3L11 11M11 3L3 11" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}
