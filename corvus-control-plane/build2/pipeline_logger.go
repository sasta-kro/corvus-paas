package build2

import (
	"fmt"
	"os"
	"time"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// deployerPipelineLogger is a helper struct for the DeployerPipeline that
// provides per-deployment logging and failure handling for a single pipeline execution.
// It writes simultaneously to the application's structured logger (main slog) and
// a deployment-specific log file (raw text). This should be constructed at the start
// of each pipeline method (eg, DeployZipUpload , Redeploy...).
type deployerPipelineLogger struct {
	// to access the app's main logger (pipeline.logger) and
	// pipeline.database (just in the logFailureAndUpdateStatus method to update the deployment status to "failed")
	pipeline   *DeployerPipeline
	deployment *models.Deployment // for .Slug and .ID
	logFile    *os.File           // nil if the log file could not be opened
}

// logInfo() writes a timestamped entry to the deployment log file and a structured
// entry to the application logger. Safe to call even if logFile is nil.
// Used throughout the deployerPipeline to record each step's outcome.
func (pipelineLogger *deployerPipelineLogger) logInfo(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), message)

	pipelineLogger.pipeline.logger.Info("deployer pipeline",
		"slug", pipelineLogger.deployment.Slug,
		"msg", message,
	)
	if pipelineLogger.logFile != nil {
		pipelineLogger.logFile.WriteString(line)
	}
}

// logFailureAndUpdateStatus() logs the failure reason to both the specific deployment log file
// and the structured application logger (slog), then updates the deployment's status to
// "failed" in the database. If the database status update itself fails, the error
// is logged to the structured logger but does not propagate further, since no
// recovery action is possible at that point.
// This is called at any pipeline step that cannot be recovered from.
// > this function is the only reason deployerPipelineLogger need access to pipeline.database
func (pipelineLogger *deployerPipelineLogger) logFailureAndUpdateStatus(reason string, err error) {
	pipelineLogger.logInfo("FAILED: %s: %v", reason, err)

	dbErr := pipelineLogger.pipeline.database.UpdateStatus(pipelineLogger.deployment.ID, models.StatusFailed)
	if dbErr != nil {
		pipelineLogger.pipeline.logger.Error("failed to update status to failed",
			"id", pipelineLogger.deployment.ID,
			"error", dbErr,
		)
	}
}
