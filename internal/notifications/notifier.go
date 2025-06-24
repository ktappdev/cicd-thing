package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/deployment"
)

// Notifier handles sending notifications
type Notifier struct {
	config *config.Config
}

// New creates a new notifier instance
func New(cfg *config.Config) *Notifier {
	return &Notifier{
		config: cfg,
	}
}

// NotifyDeploymentResult sends notifications based on deployment results
func (n *Notifier) NotifyDeploymentResult(result *deployment.Result) {
	// Only notify on specific conditions
	shouldNotify := false
	
	switch result.Status {
	case deployment.StatusFailed:
		shouldNotify = true
	case deployment.StatusRollback:
		shouldNotify = n.config.NotifyOnRollback
	case deployment.StatusTimeout:
		shouldNotify = true
	case deployment.StatusSuccess:
		// Could add config option for success notifications
		shouldNotify = false
	}

	if !shouldNotify {
		return
	}

	// Send notifications (placeholder for now)
	n.sendLogNotification(result)
	
	// Future: Add email, Slack, webhook notifications
	// n.sendEmailNotification(result)
	// n.sendSlackNotification(result)
	// n.sendWebhookNotification(result)
}

// sendLogNotification sends a log-based notification
func (n *Notifier) sendLogNotification(result *deployment.Result) {
	message := n.formatNotificationMessage(result)
	fmt.Printf("NOTIFICATION: %s\n", message)
}

// formatNotificationMessage creates a formatted notification message
func (n *Notifier) formatNotificationMessage(result *deployment.Result) string {
	switch result.Status {
	case deployment.StatusFailed:
		return fmt.Sprintf("🚨 Deployment FAILED for %s (%s): %s", 
			result.Request.Repository, result.Request.Branch, result.Error)
	case deployment.StatusRollback:
		return fmt.Sprintf("🔄 Deployment ROLLED BACK for %s (%s): %s", 
			result.Request.Repository, result.Request.Branch, result.Error)
	case deployment.StatusTimeout:
		return fmt.Sprintf("⏰ Deployment TIMED OUT for %s (%s) after %v", 
			result.Request.Repository, result.Request.Branch, result.Duration)
	case deployment.StatusSuccess:
		return fmt.Sprintf("✅ Deployment SUCCESS for %s (%s) in %v", 
			result.Request.Repository, result.Request.Branch, result.Duration)
	default:
		return fmt.Sprintf("📋 Deployment %s for %s (%s)", 
			result.Status, result.Request.Repository, result.Request.Branch)
	}
}

// SlackWebhookPayload represents a Slack webhook payload
type SlackWebhookPayload struct {
	Text        string       `json:"text"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	Fields    []Field `json:"fields,omitempty"`
	Timestamp int64   `json:"ts"`
}

// Field represents a Slack attachment field
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// sendSlackNotification sends a notification to Slack (placeholder implementation)
func (n *Notifier) sendSlackNotification(result *deployment.Result) error {
	// This would require SLACK_WEBHOOK_URL in config
	webhookURL := "" // n.config.SlackWebhookURL
	if webhookURL == "" {
		return nil // Slack not configured
	}

	color := "good"
	switch result.Status {
	case deployment.StatusFailed, deployment.StatusTimeout:
		color = "danger"
	case deployment.StatusRollback:
		color = "warning"
	}

	payload := SlackWebhookPayload{
		Username:  "CI/CD Thing",
		IconEmoji: ":robot_face:",
		Attachments: []Attachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("Deployment %s", result.Status),
				Text:      n.formatNotificationMessage(result),
				Timestamp: result.EndTime.Unix(),
				Fields: []Field{
					{Title: "Repository", Value: result.Request.Repository, Short: true},
					{Title: "Branch", Value: result.Request.Branch, Short: true},
					{Title: "Commit", Value: result.Request.Commit[:8], Short: true},
					{Title: "Duration", Value: result.Duration.String(), Short: true},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// EmailNotification represents an email notification
type EmailNotification struct {
	To      string
	Subject string
	Body    string
}

// sendEmailNotification sends an email notification (placeholder implementation)
func (n *Notifier) sendEmailNotification(result *deployment.Result) error {
	// This would require email configuration (SMTP settings)
	// For now, just return nil as it's not implemented
	
	email := EmailNotification{
		To:      "", // n.config.NotificationEmail
		Subject: fmt.Sprintf("Deployment %s: %s", result.Status, result.Request.Repository),
		Body:    n.formatEmailBody(result),
	}

	// TODO: Implement actual email sending
	fmt.Printf("EMAIL NOTIFICATION: %+v\n", email)
	return nil
}

// formatEmailBody creates an email body for the notification
func (n *Notifier) formatEmailBody(result *deployment.Result) string {
	body := fmt.Sprintf(`
Deployment Notification

Repository: %s
Branch: %s
Commit: %s
Status: %s
Duration: %s
Timestamp: %s

`, result.Request.Repository, result.Request.Branch, result.Request.Commit,
		result.Status, result.Duration, result.EndTime.Format(time.RFC3339))

	if result.Error != "" {
		body += fmt.Sprintf("Error: %s\n\n", result.Error)
	}

	if result.Output != "" {
		body += fmt.Sprintf("Output:\n%s\n", result.Output)
	}

	return body
}

// sendWebhookNotification sends a generic webhook notification (placeholder implementation)
func (n *Notifier) sendWebhookNotification(result *deployment.Result) error {
	// This would require NOTIFICATION_WEBHOOK_URL in config
	webhookURL := "" // n.config.NotificationWebhookURL
	if webhookURL == "" {
		return nil // Webhook not configured
	}

	payload := map[string]interface{}{
		"event":      "deployment",
		"status":     result.Status,
		"repository": result.Request.Repository,
		"branch":     result.Request.Branch,
		"commit":     result.Request.Commit,
		"duration":   result.Duration.Seconds(),
		"timestamp":  result.EndTime.Unix(),
		"error":      result.Error,
		"manual":     result.Request.Manual,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
