package build

import (
	"context"
	"log/slog"
	"time"
)

// StartExpirationCleanupLoop runs a background loop that checks for expired
// deployments every tickInterval and cleans them up (stop container, remove
// files, remove log, delete DB row). This is the same teardown sequence as
// the DELETE /api/deployments/:uuid handler.
//
// The loop runs until the provided context is canceled (on graceful shutdown).
// It should be launched as a goroutine from main.go.
func (deployerPipeline *DeployerPipeline) StartExpirationCleanupLoop(
	expirationContext context.Context,
	tickInterval time.Duration,
	logger *slog.Logger,
) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	logger.Info("expiration cleanup loop started", "interval", tickInterval.String())

	for {
		select {
		case <-expirationContext.Done():
			logger.Info("expiration cleanup loop stopped")
			return
		case <-ticker.C:
			deployerPipeline.cleanupExpiredDeployments(expirationContext, logger)
		}
	}
}

// cleanupExpiredDeployments fetches all expired live deployments and runs
// the full teardown for each one. Errors on individual deployments are
// logged but do not stop the loop from processing the remaining ones.
func (deployerPipeline *DeployerPipeline) cleanupExpiredDeployments(
	cleanupDeploymentsContext context.Context,
	logger *slog.Logger,
) {
	expiredDeployments, err := deployerPipeline.database.ListExpiredDeployments()
	if err != nil {
		logger.Error("failed to list expired deployments", "error", err)
		return
	}

	if len(expiredDeployments) == 0 {
		return
	}

	logger.Info("found expired deployments", "count", len(expiredDeployments))

	for _, deployment := range expiredDeployments {
		logger.Info("cleaning up expired deployment",
			"id", deployment.ID,
			"slug", deployment.Slug,
			"expires_at", deployment.ExpiresAt,
		)

		if err := deployerPipeline.TeardownDeployment(cleanupDeploymentsContext, deployment); err != nil {
			logger.Error("failed to teardown/remove expired deployment",
				"id", deployment.ID,
				"slug", deployment.Slug,
				"error", err,
			)
			continue
		}

		logger.Info("expired deployment cleaned up",
			"id", deployment.ID,
			"slug", deployment.Slug,
		)
	}
}
