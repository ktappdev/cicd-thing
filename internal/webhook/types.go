package webhook

import "time"

// GitHubWebhookPayload represents the GitHub webhook payload structure
type GitHubWebhookPayload struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	HeadCommit HeadCommit `json:"head_commit"`
	Commits    []Commit   `json:"commits"`
}

// Repository represents the repository information in the webhook payload
type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
}

// Pusher represents the user who pushed the commits
type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// HeadCommit represents the head commit information
type HeadCommit struct {
	ID        string    `json:"id"`
	TreeID    string    `json:"tree_id"`
	Distinct  bool      `json:"distinct"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    Author    `json:"author"`
	Committer Author    `json:"committer"`
	Added     []string  `json:"added"`
	Removed   []string  `json:"removed"`
	Modified  []string  `json:"modified"`
}

// Commit represents individual commit information
type Commit struct {
	ID        string    `json:"id"`
	TreeID    string    `json:"tree_id"`
	Distinct  bool      `json:"distinct"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    Author    `json:"author"`
	Committer Author    `json:"committer"`
	Added     []string  `json:"added"`
	Removed   []string  `json:"removed"`
	Modified  []string  `json:"modified"`
}

// Author represents commit author/committer information
type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// DeploymentRequest represents a deployment request extracted from webhook
type DeploymentRequest struct {
	Repository string
	Branch     string
	Commit     string
	Message    string
	Author     string
	Timestamp  time.Time
	LocalPath  string
}
