package build

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZipUpload extracts the contents of a zip archive at zipFilePath into the
// destination directory `destinationDirectory`. the dest dir is created if it does not exist.
// (security) zip files can contain paths with ".." components (zip slip attack),
// which would allow a malicious archive to write files outside the destinationDirectory.
// Every extracted path is validated to ensure it stays within destinationDirectory before
// any file or directory is created on disk.
func ExtractZipUpload(zipFilePath string, destinationDirectory string) error {
	// os.MkdirAll creates destinationDirectory and any missing parents.
	// chmod 0755 - owner has read/write/execute, group and others have read/execute.
	// (execute permission on a directory means the ability to cd into it)
	errMakeDir := os.MkdirAll(destinationDirectory, 0755)
	if errMakeDir != nil {
		return fmt.Errorf("failed to create extraction directory %q: %w", destinationDirectory, errMakeDir)
	}

	// zip.OpenReader opens the zip archive for reading.
	// it reads the central directory at the end of the file to get the list of entries.
	// this is more efficient than scanning the file from the start.
	zipReader, errOpenZip := zip.OpenReader(zipFilePath)
	if errOpenZip != nil {
		return fmt.Errorf("failed to open zip archive %q: %w", zipFilePath, errOpenZip)
	}
	defer zipReader.Close() // zip reader returns a ReaderCloser so it must be closed

	// zip entry is just a file/folder in a zip file
	for _, zipEntry := range zipReader.File {
		errExtractZipEntry := extractZipEntry(zipEntry, destinationDirectory)
		if errExtractZipEntry != nil { // helper func
			return fmt.Errorf("failed to extract entry %q: %w", zipEntry.Name, errExtractZipEntry)
		}
	}

	return nil // function just write files, doesnt return anything except an error
}

// extractZipEntry (helper func) extracts a single file or directory entry from the zip archive.
// Separated from ExtractZipUpload to keep the loop body readable and to allow
// proper defer handling. `defers` inside a loop body do not run until the
// function returns/end, which means file handles stay open for the entire loop.
// by moving the logic into a called function, each entry's file handle is
// deferred and closed before the next entry is processed.
func extractZipEntry(zipEntry *zip.File, destinationDirectory string) error {
	// filepath.Join cleans the path, resolving any `..` components (cmd to go to parent folder)
	// The security check below then verifies the cleaned path still starts
	// with destinationDirectory, rejecting any entry that tried to escape via `..`
	entryDestPath := filepath.Join(destinationDirectory, zipEntry.Name)

	// Zip slip exploit protection (security) Filepath.Clean resolves `..` segments.
	// example malicious entry name: "../../etc/passwd"
	// after Join + Clean: "/tmp/builds/abc123/../../etc/passwd" -> "/etc/passwd"
	// the HasPrefix() check catches this and rejects it.
	safePrefix := filepath.Clean(destinationDirectory) + string(os.PathSeparator) // well tbh .Clean() is not necessary here (cuz hard coded)

	// but .Clean() here is necessary cuz zip entry might have sus file names
	potentiallyMaliciousEntryDestPath := filepath.Clean(entryDestPath) + string(os.PathSeparator)
	if !strings.HasPrefix(potentiallyMaliciousEntryDestPath, safePrefix) {
		// this shouldn't really be reached if the path is cleaned properly.
		// but sometimes, even after cleaning, the file might resolve to a path outside the intended dir. so error for that
		return fmt.Errorf("zip slip detected: entry %q would write outside destination directory", zipEntry.Name)
	}

	// if folder, just create folder with os.MkdirAll() and bubble up the error if happens
	if zipEntry.FileInfo().IsDir() {
		// create the directory with standard permissions (owner = read/write/execute, group & others = read/execute)
		// the trailing separator is included in the prefix check above,
		// so directories are also validated before creation.
		return os.MkdirAll(entryDestPath, 0755)
	}

	// creating the parent directory of this file entry.
	// since zip archives can contain files without explicit directory entries,
	// so the parent directory may not have been created yet.
	errMakeParentFolder := os.MkdirAll(filepath.Dir(entryDestPath), 0755)
	if errMakeParentFolder != nil {
		return fmt.Errorf("failed to create parent directory for %q: %w", entryDestPath, errMakeParentFolder)
	}

	return writeZipEntryToDisk(zipEntry, entryDestPath) // helper
}

// writeZipEntryToDisk opens the zip entry for reading and writes its contents
// to the destination path on disk. File permissions are taken from the zip entry's stored mode, falling back
// to 0644 (owner read/write, group and others read-only) if the mode is zero (happens if zipped in Windows).
// 0644 is the standard permission for files that should be readable but not executable.
func writeZipEntryToDisk(zipEntry *zip.File, destinationPath string) error {
	// open the zip entry's compressed data stream for reading
	zipEntryReadCloser, errOpenZipEntry := zipEntry.Open()
	if errOpenZipEntry != nil {
		return fmt.Errorf("failed to open zip entry for reading: %w", errOpenZipEntry)
	}
	defer zipEntryReadCloser.Close() // gotta close after reading

	// Checking file permissions from the zip entry metadata.
	// zip entries built on Unix systems store the original file mode. But entries created on
	// Windows often store mode 0, so the fallback ensures the extracted file is always readable.
	filePermissionsMode := zipEntry.Mode()
	if filePermissionsMode == 0 {
		filePermissionsMode = 0644
	}

	// os.OpenFile() creates the destination file (truncates = overrides any existing file with same name)
	// os.O_CREATE - create the file if it does not exist.
	// os.O_WRONLY - open for writing only no reading (the zip entry is the source, not this file).
	// os.O_TRUNC - Clears all existing data (truncates to 0 bytes) if the file already exists,
	// preventing data corruption from previous extractions.
	// The pipe character `|` is a bitwise OR operator. It merges the individual binary flags
	// into a single configuration payload sent to the operating system.
	destinationFile, errOpenZipEntry := os.OpenFile(
		destinationPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		filePermissionsMode,
	)
	if errOpenZipEntry != nil {
		return fmt.Errorf("failed to create destination file %q: %w", destinationPath, errOpenZipEntry)
	}
	defer destinationFile.Close()
	// this is a pointer (open file handle) on the host filesystem. The defer call makes the os
	// releases the file descriptor and flushes all pending data to the disk after the function execution ends.

	// io.Copy() decompresses the zip entry and streams the decompressed entry content
	// into the destination file. It reads data as chunks internally, so arbitrarily large
	// files are handled without loading the entire entry into memory at once.
	_, errUncompressAndCopy := io.Copy(destinationFile, zipEntryReadCloser)
	if errUncompressAndCopy != nil {
		return fmt.Errorf("failed to write zip entry content to disk: %w", errUncompressAndCopy)
	}

	return nil
}
