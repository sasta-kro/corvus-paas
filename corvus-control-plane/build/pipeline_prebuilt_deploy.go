package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sasta-kro/corvus-paas/corvus-control-plane/models"
)

// DeployPrebuilt is the pipeline for deploying from pre-built static files
// stored on the server's filesystem. Used for quick-deploy presets (Vite Starter,
// React App, etc.) to skip the clone + build steps entirely.
//
// The preset files live at <presetStorageRoot>/<presetID>/ and must contain
// an index.html at their root. The pipeline copies these files to the
// deployment's asset directory and starts an Nginx container to serve them.
//
// For the "Your Message" preset, the pipeline performs a string replacement
// of {{CORVUS_MESSAGE}} in the copied index.html with the user-provided message
// from the deployment's environment variables.
//
// Called as a goroutine from the handler: go pipeline.DeployPrebuilt(deployment)
func (deployerPipeline *DeployerPipeline) DeployPrebuilt(deployment *models.Deployment) {
	deployContext := context.Background()

	// ===== Open log file and create pipeline logger
	logFile, errOpenLogFile := deployerPipeline.openLogFileForCurrentDeployment(deployment.Slug)
	if errOpenLogFile != nil {
		deployerPipeline.logger.Error("failed to open deployment log file",
			"slug", deployment.Slug,
			"error", errOpenLogFile,
		)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	pipelineLogger := &deployerPipelineLogger{
		pipeline:   deployerPipeline,
		deployment: deployment,
		logFile:    logFile,
	}

	// ===== Set status to deploying
	pipelineLogger.logInfo("starting prebuilt deployment pipeline (preset: %s)", safePresetID(deployment.PresetID))
	statusError := deployerPipeline.database.UpdateStatus(deployment.ID, models.StatusDeploying)
	if statusError != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to set status to deploying", statusError)
		return
	}

	// ===== Validate preset ID and resolve source directory
	if deployment.PresetID == nil || *deployment.PresetID == "" {
		pipelineLogger.logFailureAndUpdateStatus("preset_id is required for prebuilt deployments",
			fmt.Errorf("missing preset_id on deployment %q", deployment.ID),
		)
		return
	}

	presetID := *deployment.PresetID
	presetSourceDir := filepath.Join(deployerPipeline.presetStorageRoot, presetID)

	// verify the preset directory exists on disk
	_, errStat := os.Stat(presetSourceDir)
	if os.IsNotExist(errStat) {
		pipelineLogger.logFailureAndUpdateStatus(
			fmt.Sprintf("preset %q not found at %s", presetID, presetSourceDir),
			fmt.Errorf("preset directory does not exist: %s", presetSourceDir),
		)
		return
	}
	if errStat != nil {
		pipelineLogger.logFailureAndUpdateStatus("failed to stat preset directory", errStat)
		return
	}

	pipelineLogger.logInfo("preset source directory found: %s", presetSourceDir)

	// ===== Handle dynamic message injection for "your-message" preset
	// The "Your Message" preset has a {{CORVUS_MESSAGE}} placeholder in its index.html
	// that gets replaced with the user's custom text. This is done after copying
	// to the asset directory so the original preset files remain untouched.
	needsMessageInjection := false
	if presetID == "your-message" {
		envVarsList, envDecodeError := decodeEnvVarsToSlice(deployment.EnvironmentVariables)
		if envDecodeError == nil {
			for _, envVar := range envVarsList {
				if strings.HasPrefix(envVar, "VITE_CORVUS_MESSAGE=") {
					needsMessageInjection = true
					break
				}
			}
		}
	}

	// ===== Deploy to Nginx using the preset source directory
	// The output directory for prebuilt presets is always "." (the preset root).
	// Override it so deployToNginx copies from the correct location.
	deployment.OutputDirectory = "."

	// deployToNginx handles: copy to asset storage, stop existing container, start nginx, set status live.
	success := deployerPipeline.deployToNginx(
		deployContext,
		deployment,
		presetSourceDir,
		pipelineLogger,
	)

	// ===== Post-deploy: inject custom message if needed
	if success && needsMessageInjection {
		deployerPipeline.injectCustomMessage(deployment, pipelineLogger)
	}
}

// injectCustomMessage reads the deployed index.html, replaces the
// {{CORVUS_MESSAGE}} placeholder with the user's message, and writes it back.
// This runs after deployToNginx has already copied files to the asset storage root.
func (deployerPipeline *DeployerPipeline) injectCustomMessage(
	deployment *models.Deployment,
	pipelineLogger *deployerPipelineLogger,
) {
	envVarsList, err := decodeEnvVarsToSlice(deployment.EnvironmentVariables)
	if err != nil {
		pipelineLogger.logInfo("WARNING: could not decode env vars for message injection: %v", err)
		return
	}

	var message string
	for _, envVar := range envVarsList {
		if strings.HasPrefix(envVar, "VITE_CORVUS_MESSAGE=") {
			message = strings.TrimPrefix(envVar, "VITE_CORVUS_MESSAGE=")
			break
		}
	}

	if message == "" {
		pipelineLogger.logInfo("no VITE_CORVUS_MESSAGE found, skipping message injection")
		return
	}

	indexPath := filepath.Join(deployerPipeline.assetStorageRoot, deployment.Slug, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		pipelineLogger.logInfo("WARNING: could not read index.html for message injection: %v", err)
		return
	}

	updated := strings.ReplaceAll(string(content), "{{CORVUS_MESSAGE}}", message)
	err = os.WriteFile(indexPath, []byte(updated), 0644)
	if err != nil {
		pipelineLogger.logInfo("WARNING: could not write index.html after message injection: %v", err)
		return
	}

	pipelineLogger.logInfo("custom message injected into index.html")
}

// safePresetID returns the preset ID string or "<nil>" if the pointer is nil.
// used for logging to avoid nil pointer dereference.
func safePresetID(presetID *string) string {
	if presetID == nil {
		return "<nil>"
	}
	return *presetID
}
