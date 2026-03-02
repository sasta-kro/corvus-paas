import { useState, useRef, useCallback } from "react";
import { formatFileSize } from "../../lib/utils";

interface DragDropZoneProps {
  onFileSelected: (file: File) => void;
  onFileRemoved: () => void;
  selectedFile: File | null;
  error: string | null;
  disabled: boolean;
}

export default function DragDropZone({
  onFileSelected, onFileRemoved, selectedFile, error, disabled,
}: DragDropZoneProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const dragCounterRef = useRef(0);
  const [isDragOver, setIsDragOver] = useState(false);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault(); e.stopPropagation();
    dragCounterRef.current++;
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault(); e.stopPropagation();
    dragCounterRef.current--;
    if (dragCounterRef.current === 0) setIsDragOver(false);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault(); e.stopPropagation();
  }, []);

  const validateAndSelect = useCallback((file: File) => onFileSelected(file), [onFileSelected]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault(); e.stopPropagation();
    dragCounterRef.current = 0;
    setIsDragOver(false);
    if (disabled) return;
    const files = e.dataTransfer.files;
    if (files.length > 0) validateAndSelect(files[0]);
  }, [disabled, validateAndSelect]);

  const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) validateAndSelect(files[0]);
    e.target.value = "";
  }, [validateAndSelect]);

  if (selectedFile) {
    return (
      <div
        className="p-5"
        style={{
          border: "none",
          borderRadius: "3px 5px 4px 3px",
          boxShadow: "2px 3px 0 -1px rgba(53,51,48,0.12), 0 1px 0 0 rgba(17,17,16,0.04)",
          background: "var(--paper-warm)",
        }}
      >
        <div className="flex items-center justify-between">
          <div>
            <p style={{ color: "var(--sumi)", fontWeight: 700, fontSize: "0.95rem" }}>{selectedFile.name}</p>
            <p style={{ color: "var(--sumi-wash)", fontSize: "0.8rem" }}>{formatFileSize(selectedFile.size)}</p>
          </div>
          <button
            onClick={onFileRemoved}
            disabled={disabled}
            className="cursor-pointer transition-colors"
            style={{ color: "var(--sumi-wash)" }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "var(--vermillion)")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "var(--sumi-wash)")}
          >
            <svg width="16" height="16" viewBox="0 0 14 14" fill="none">
              <path d="M3 3L11 11M11 3L3 11" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
            </svg>
          </button>
        </div>
        {error && <p style={{ color: "var(--vermillion)", fontSize: "0.8rem", marginTop: "0.5rem" }}>{error}</p>}
      </div>
    );
  }

  return (
    <div
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      onClick={() => !disabled && inputRef.current?.click()}
      className="p-12 text-center transition-all"
      style={{
        border: "none",
        borderRadius: "3px 5px 4px 3px",
        background: isDragOver ? "var(--paper-warm)" : "transparent",
        backgroundImage: isDragOver
          ? "none"
          : `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100%25' height='100%25'%3E%3Crect width='100%25' height='100%25' fill='none' stroke='%23b5b0a3' stroke-width='3' stroke-dasharray='12 8 4 8' stroke-dashoffset='0' stroke-linecap='round'/%3E%3C/svg%3E")`,
        boxShadow: isDragOver
          ? "inset 0 0 20px rgba(17,17,16,0.06), 3px 4px 0 -1px rgba(17,17,16,0.08)"
          : "none",
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.5 : 1,
      }}
    >
      {/* Small ink drop icon */}
      <svg width="28" height="28" viewBox="0 0 24 24" fill="var(--sumi-ghost)" className="mx-auto mb-3">
        <path d="M12 2C12 2 6 10 6 14a6 6 0 0 0 12 0c0-4-6-12-6-12zM12 18a4 4 0 0 1-4-4c0-1 0.3-2.2 1-3.5l3-5 3 5c0.7 1.3 1 2.5 1 3.5a4 4 0 0 1-4 4z" />
      </svg>
      <p style={{ color: "var(--sumi-light)", fontSize: "0.95rem" }}>
        {isDragOver ? "Drop your file here" : "Drag and drop a .zip file, or click to browse"}
      </p>
      <p style={{ color: "var(--sumi-ghost)", fontSize: "0.8rem", marginTop: "0.3rem", fontStyle: "italic" }}>
        Maximum file size: 50MB
      </p>
      <input ref={inputRef} type="file" accept=".zip" onChange={handleFileInput} className="hidden" aria-label="Upload zip file" />
      {error && <p style={{ color: "var(--vermillion)", fontSize: "0.8rem", marginTop: "0.75rem" }}>{error}</p>}
    </div>
  );
}
