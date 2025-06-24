package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/deployment"
	"github.com/ktappdev/cicd-thing/internal/logger"
	"github.com/ktappdev/cicd-thing/internal/notifications"
	"github.com/ktappdev/cicd-thing/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	deployLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer deployLogger.Close()

	deployLogger.LogInfo("Starting CI/CD Thing deployment orchestrator")

	// Initialize notification system
	notifier := notifications.New(cfg)
	deployLogger.LogInfo("Notification system initialized")

	// Initialize deployment executor
	executor := deployment.New(cfg)
	deployLogger.LogInfo("Deployment executor initialized")

	// Start deployment result processor
	go processDeploymentResults(executor, deployLogger, notifier)

	// Create and start the server
	srv := server.New(cfg, executor, deployLogger)
	deployLogger.LogInfo("Server initialized")

	// Handle graceful shutdown
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	deployLogger.LogInfo("Shutting down CI/CD Thing deployment orchestrator")
}

// processDeploymentResults processes deployment results and logs them
func processDeploymentResults(executor *deployment.Executor, deployLogger *logger.Logger, notifier *notifications.Notifier) {
	for result := range executor.GetResults() {
		deployLogger.LogDeploymentResult(result)

		// Send notifications
		notifier.NotifyDeploymentResult(result)

		// Log additional info based on status
		switch result.Status {
		case deployment.StatusSuccess:
			deployLogger.LogInfo("Deployment completed successfully for " + result.Request.Repository)
		case deployment.StatusFailed:
			deployLogger.LogError("Deployment failed for "+result.Request.Repository, nil)
		case deployment.StatusTimeout:
			deployLogger.LogError("Deployment timed out for "+result.Request.Repository, nil)
		case deployment.StatusRollback:
			deployLogger.LogInfo("Deployment rolled back for " + result.Request.Repository)
		}
	}
}
