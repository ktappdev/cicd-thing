package mapping

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ktappdev/cicd-thing/internal/config"
)

// Mapper handles repository to local path mapping
type Mapper struct {
	config *config.Config
}

// New creates a new mapper instance
func New(cfg *config.Config) *Mapper {
	return &Mapper{
		config: cfg,
	}
}

// GetLocalPath returns the local path for a given repository
func (m *Mapper) GetLocalPath(repoFullName string) (string, error) {
	localPath, exists := m.config.RepoMap[repoFullName]
	if !exists {
		return "", fmt.Errorf("no mapping found for repository %s", repoFullName)
	}

	// Expand path
	expandedPath, err := m.expandPath(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to expand path %s: %w", localPath, err)
	}

	return expandedPath, nil
}

// GetAppName extracts the application name from repository full name
// For example: "octocat/Hello-World" -> "Hello-World"
func (m *Mapper) GetAppName(repoFullName string) string {
	parts := strings.Split(repoFullName, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return repoFullName
}

// expandPath expands ~ and environment variables in paths
func (m *Mapper) expandPath(path string) (string, error) {
	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand tilde
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// ValidatePath checks if a local path exists and is accessible
func (m *Mapper) ValidatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("failed to access path %s: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if we can read the directory
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read directory %s: %w", path, err)
	}
	file.Close()

	return nil
}

// ListMappings returns all configured repository mappings
func (m *Mapper) ListMappings() map[string]string {
	result := make(map[string]string)
	for repo, path := range m.config.RepoMap {
		expandedPath, err := m.expandPath(path)
		if err != nil {
			// If expansion fails, use original path
			expandedPath = path
		}
		result[repo] = expandedPath
	}
	return result
}
