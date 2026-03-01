import { useState, useCallback } from "react";
import DragDropZone from "./DragDropZone";
import DeployButton from "./DeployButton";
import { MAX_FILE_SIZE_BYTES } from "../../config/constants";

interface ZipUploadTabProps {
  onDeploy: (file: File, outputDirectory: string, buildCommand: string) => void;
  disabled: boolean;
}

/** Zip Upload tab — drag-drop zone + config fields */
export default function ZipUploadTab({ onDeploy, disabled }: ZipUploadTabProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [outputDirectory, setOutputDirectory] = useState(".");
  const [buildCommand, setBuildCommand] = useState("");

  const handleFileSelected = useCallback((file: File) => {
    setFileError(null);
    if (!file.name.endsWith(".zip")) {
      setFileError("Only .zip files are accepted.");
      setSelectedFile(null);
      return;
    }
    if (file.size > MAX_FILE_SIZE_BYTES) {
      setFileError("File exceeds the 50MB limit.");
      setSelectedFile(null);
      return;
    }
    setSelectedFile(file);
  }, []);

  const handleFileRemoved = useCallback(() => {
    setSelectedFile(null);
    setFileError(null);
  }, []);

  const handleDeploy = () => {
    if (selectedFile) {
      onDeploy(selectedFile, outputDirectory, buildCommand);
    }
  };

  return (
    <div className="space-y-4">
      <DragDropZone
        onFileSelected={handleFileSelected}
        onFileRemoved={handleFileRemoved}
        selectedFile={selectedFile}
        error={fileError}
        disabled={disabled}
      />

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label htmlFor="zip-output-dir" className="block text-sm font-medium mb-1">
            Output Directory
          </label>
          <input
            id="zip-output-dir"
            type="text"
            value={outputDirectory}
            onChange={(e) => setOutputDirectory(e.target.value)}
            placeholder="e.g., dist, build, ."
            disabled={disabled}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
          />
        </div>
        <div>
          <label htmlFor="zip-build-cmd" className="block text-sm font-medium mb-1">
            Build Command (optional)
          </label>
          <input
            id="zip-build-cmd"
            type="text"
            value={buildCommand}
            onChange={(e) => setBuildCommand(e.target.value)}
            placeholder="e.g., npm ci && npm run build"
            disabled={disabled}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-black disabled:bg-gray-50"
          />
        </div>
      </div>

      <DeployButton
        onClick={handleDeploy}
        disabled={disabled || !selectedFile || !!fileError}
        loading={disabled}
      />
    </div>
  );
}

