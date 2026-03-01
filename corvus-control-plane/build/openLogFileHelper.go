package build

import (
	"fmt"
	"os"
	"path/filepath"
)

// openLogFile creates or opens the log file for a deployment (each deployment has its own log file).
// the log directory is created if it does not exist.
// the file is opened in append mode so redeployments add to the existing log
// rather than overwriting it, preserving the full deployment history in one file.
func (deployerPipeline *DeployerPipeline) openLogFile(slug string) (*os.File, error) {
	err := os.MkdirAll(deployerPipeline.logRoot, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	logPath := filepath.Join(deployerPipeline.logRoot, slug+".log")
	// os.O_APPEND: writes go to the end of the file.
	// os.O_CREATE: create the file if it does not exist.
	// os.O_WRONL: open for writing only.
	// 0644: owner read/write, group and others read-only.
	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
