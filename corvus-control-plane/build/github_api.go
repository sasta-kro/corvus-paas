package build

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// githubRepoInfo holds the subset of fields returned by the GitHub API
// for a public repository. Only the fields we actually use are declared;
// json.Unmarshal silently ignores everything else.
type githubRepoInfo struct {
	DefaultBranch string `json:"default_branch"`
}

// checkGitHubRepoExists verifies that a public GitHub repository is accessible
// before attempting a potentially hanging git clone. Returns nil if the repo
// exists and is publicly accessible, or an error with a clear message if not.
//
// This prevents the pipeline from hanging indefinitely when a repo URL is
// invalid, private, or points to a deleted repository.
func checkGitHubRepoExists(repoURL string, logger *slog.Logger) error {
	ownerAndRepo, err := extractOwnerRepo(repoURL)
	if err != nil {
		return fmt.Errorf("invalid GitHub URL %q: %w", repoURL, err)
	}

	apiURL := "https://api.github.com/repos/" + ownerAndRepo

	logger.Info("checking if github repository exists", "url", repoURL, "api_url", apiURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to reach GitHub API to verify repo %q: %w", ownerAndRepo, err)
	}
	defer resp.Body.Close()
	// drain the body so the connection can be reused
	io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		logger.Info("github repository exists and is accessible", "repo", ownerAndRepo)
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("GitHub repository %q does not exist or is private", ownerAndRepo)
	case http.StatusForbidden:
		// 403 usually means rate limited
		logger.Warn("github API returned 403 (possible rate limit), skipping existence check", "repo", ownerAndRepo)
		return nil // optimistic: let the clone attempt proceed
	default:
		return fmt.Errorf("GitHub API returned unexpected status %d for %q", resp.StatusCode, ownerAndRepo)
	}
}

// fetchGitHubDefaultBranch queries the GitHub API to determine the default
// branch of a public repository. The repoURL is expected to be a standard
// GitHub clone URL like "https://github.com/user/repo.git".
//
// Returns the default branch name (e.g. "main", "master", "develop") or
// an error if the API call fails or the URL cannot be parsed.
//
// This is used to auto-correct the branch when the user left the default
// "main" but the repo actually uses "master" or something else.
func fetchGitHubDefaultBranch(repoURL string) (string, error) {
	// extract "user/repo" from the clone URL
	// handles both "https://github.com/user/repo.git" and "https://github.com/user/repo"
	ownerAndRepo, err := extractOwnerRepo(repoURL)
	if err != nil {
		return "", err
	}

	apiURL := "https://api.github.com/repos/" + ownerAndRepo

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("github API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned status %d for %s", resp.StatusCode, apiURL)
	}

	var info githubRepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to decode github API response: %w", err)
	}

	if info.DefaultBranch == "" {
		return "", fmt.Errorf("github API returned empty default_branch for %s", ownerAndRepo)
	}

	return info.DefaultBranch, nil
}

// extractOwnerRepo parses a GitHub URL and returns "owner/repo".
// accepts formats like:
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo/
func extractOwnerRepo(repoURL string) (string, error) {
	// strip trailing slash and .git suffix
	cleaned := strings.TrimRight(repoURL, "/")
	cleaned = strings.TrimSuffix(cleaned, ".git")

	// find the github.com part and extract what comes after
	const prefix = "github.com/"
	idx := strings.Index(cleaned, prefix)
	if idx == -1 {
		return "", fmt.Errorf("not a github.com URL: %s", repoURL)
	}

	ownerRepo := cleaned[idx+len(prefix):]
	parts := strings.Split(ownerRepo, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("could not extract owner/repo from URL: %s", repoURL)
	}

	return parts[0] + "/" + parts[1], nil
}
