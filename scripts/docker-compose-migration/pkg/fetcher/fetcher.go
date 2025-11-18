package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubFetcher fetches files from a GitHub repository
type GitHubFetcher struct {
	RepoURL    string
	Branch     string
	DockerPath string
	client     *http.Client
}

// NewGitHubFetcher creates a new GitHubFetcher instance
func NewGitHubFetcher(repoURL, branch, dockerPath string) *GitHubFetcher {
	return &GitHubFetcher{
		RepoURL:    strings.TrimSuffix(repoURL, "/"),
		Branch:     branch,
		DockerPath: strings.Trim(dockerPath, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchFiles fetches docker-compose.yml and .env.example from GitHub
func (f *GitHubFetcher) FetchFiles() (composeData []byte, envData []byte, err error) {
	// Convert GitHub URL to raw content URL
	rawBaseURL := f.getRawBaseURL()

	// Fetch docker-compose.yml
	composeURL := fmt.Sprintf("%s/%s/docker-compose.yml", rawBaseURL, f.DockerPath)
	composeData, err = f.fetchFile(composeURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch docker-compose.yml: %w", err)
	}

	// Fetch .env.example
	envURL := fmt.Sprintf("%s/%s/.env.example", rawBaseURL, f.DockerPath)
	envData, err = f.fetchFile(envURL)
	if err != nil {
		// .env.example is optional
		fmt.Printf("Warning: failed to fetch .env.example: %v\n", err)
		envData = []byte{}
	}

	return composeData, envData, nil
}

// getRawBaseURL converts a GitHub repo URL to the raw content URL
func (f *GitHubFetcher) getRawBaseURL() string {
	// Convert https://github.com/owner/repo to https://raw.githubusercontent.com/owner/repo/branch
	repoURL := strings.TrimPrefix(f.RepoURL, "https://github.com/")
	repoURL = strings.TrimPrefix(repoURL, "http://github.com/")
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s", repoURL, f.Branch)
}

// fetchFile fetches a single file from a URL
func (f *GitHubFetcher) fetchFile(url string) ([]byte, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
