package util

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyDirectory recursively copies the contents of srcDir into destDir.
// destDir is removed and recreated to prevent stale files from surviving redeploys.
// Symlinks and non-regular files (device nodes, FIFOs, sockets) are rejected
// because build output from untrusted uploads must not reference or include them.
// Note: This is designed for copying build output of static sites, Ownership, timestamps, xattrs, and ACLs are not preserved.
func CopyDirectory(srcDir string, destDir string) error {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %q: %w", srcDir, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source path %q is not a directory", srcDir)
	}

	// Wipe the destination to remove files that no longer exist in the new deploy.
	// without this, a file deleted between deploys remains served indefinitely.
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to remove destination directory %q: %w", destDir, err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %q: %w", destDir, err)
	}

	// WalkDir traverses the tree without calling os.Stat for every entry.
	// DirEntry.Type carries enough metadata to detect directories and symlinks.
	return filepath.WalkDir(srcDir, func(srcPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %q: %w", srcPath, err)
		}
		destPath := filepath.Join(destDir, relPath)

		// Reject symlinks explicitly. A symlink in untrusted build output can point outside
		// the srcDir and cause the copy to read arbitrary host files into the served directory.
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not allowed in deployment output: %q", srcPath)
		}

		if entry.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Reject device nodes, FIFOs, sockets, and any other non-regular file type.
		// Attempting to open a FIFO blocks the goroutine indefinitely.
		// Attempting to read a device node can yield arbitrary kernel data.
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported file type in deployment output: %q (type: %v)", srcPath, entry.Type())
		}

		return copyFile(srcPath, destPath)
	})
}

// copyFile copies a single regular file from src to dest.
// The destination is created or truncated if it already exists.
// File permission bits from the source are preserved on the copy.
func copyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %q: %w", src, err)
	}

	// Create or truncate the destination file with the source file mode.
	// os.O_TRUNC prevents leftover bytes if the previous file was larger.
	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dest, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content from %q to %q: %w", src, dest, err)
	}

	return nil
}
