import { useState, useRef, useCallback } from "react";
import { formatFileSize } from "../../lib/utils";

interface DragDropZoneProps {
  onFileSelected: (file: File) => void;
  onFileRemoved: () => void;
  selectedFile: File | null;
  error: string | null;
  disabled: boolean;
}

/** Drag-and-drop zone for zip file uploads */
export default function DragDropZone({
  onFileSelected,
  onFileRemoved,
  selectedFile,
  error,
  disabled,
}: DragDropZoneProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const dragCounterRef = useRef(0);
  const [isDragOver, setIsDragOver] = useState(false);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounterRef.current++;
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCounterRef.current--;
    if (dragCounterRef.current === 0) {
      setIsDragOver(false);
    }
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const validateAndSelect = useCallback(
    (file: File) => {
      onFileSelected(file);
    },
    [onFileSelected]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCounterRef.current = 0;
      setIsDragOver(false);
      if (disabled) return;
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        validateAndSelect(files[0]);
      }
    },
    [disabled, validateAndSelect]
  );

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = e.target.files;
      if (files && files.length > 0) {
        validateAndSelect(files[0]);
      }
      // Reset input so same file can be re-selected
      e.target.value = "";
    },
    [validateAndSelect]
  );

  if (selectedFile) {
    return (
      <div className="border-2 border-gray-300 border-dashed rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">{selectedFile.name}</p>
            <p className="text-xs text-gray-500">
              {formatFileSize(selectedFile.size)}
            </p>
          </div>
          <button
            onClick={onFileRemoved}
            disabled={disabled}
            className="text-gray-400 hover:text-red-500 text-xl font-bold cursor-pointer disabled:cursor-not-allowed"
          >
            ×
          </button>
        </div>
        {error && <p className="text-red-500 text-xs mt-2">{error}</p>}
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
      className={`border-2 border-dashed rounded-lg p-10 text-center transition-colors cursor-pointer ${
        disabled
          ? "border-gray-200 bg-gray-50 cursor-not-allowed"
          : isDragOver
          ? "border-black bg-gray-50"
          : "border-gray-300 hover:border-black"
      }`}
    >
      <p className="text-sm text-gray-500 mb-1">
        {isDragOver
          ? "Drop your file here"
          : "Drag and drop a .zip file here, or click to browse"}
      </p>
      <p className="text-xs text-gray-400">Maximum file size: 50MB</p>
      <input
        ref={inputRef}
        type="file"
        accept=".zip"
        onChange={handleFileInput}
        className="hidden"
        aria-label="Upload zip file"
      />
      {error && <p className="text-red-500 text-xs mt-3">{error}</p>}
    </div>
  );
}

