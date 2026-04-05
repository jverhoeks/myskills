package browse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// RepoEntry represents a unique repo from the skills.sh leaderboard.
type RepoEntry struct {
	Owner      string
	Repo       string
	SkillCount int
	Installs   int
}

// OwnerRepo returns the "owner/repo" string.
func (r RepoEntry) OwnerRepo() string {
	return r.Owner + "/" + r.Repo
}

// FetchLeaderboard fetches the skills.sh leaderboard and returns unique repos
// sorted by total installs descending.
func FetchLeaderboard() ([]RepoEntry, error) {
	// Try the API endpoint first (JSON)
	repos, err := fetchFromAPI()
	if err == nil && len(repos) > 0 {
		return repos, nil
	}

	// Fallback: scrape the HTML page
	return fetchFromHTML()
}

// fetchFromAPI tries the skills.sh API if available.
func fetchFromAPI() ([]RepoEntry, error) {
	resp, err := http.Get("https://skills.sh/api/skills")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON array of skills
	var skills []struct {
		Owner    string `json:"owner"`
		Repo     string `json:"repo"`
		Name     string `json:"name"`
		Installs int    `json:"installs"`
	}
	if err := json.Unmarshal(body, &skills); err != nil {
		return nil, err
	}

	return aggregateRepos(skills), nil
}

func aggregateRepos(skills []struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
}) []RepoEntry {
	type key struct{ owner, repo string }
	m := map[key]*RepoEntry{}

	for _, s := range skills {
		k := key{s.Owner, s.Repo}
		if e, ok := m[k]; ok {
			e.SkillCount++
			e.Installs += s.Installs
		} else {
			m[k] = &RepoEntry{
				Owner:      s.Owner,
				Repo:       s.Repo,
				SkillCount: 1,
				Installs:   s.Installs,
			}
		}
	}

	var result []RepoEntry
	for _, e := range m {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Installs > result[j].Installs
	})
	return result
}

// fetchFromHTML scrapes the skills.sh page for repo entries.
func fetchFromHTML() ([]RepoEntry, error) {
	resp, err := http.Get("https://skills.sh/")
	if err != nil {
		return nil, fmt.Errorf("fetching skills.sh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("skills.sh returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return parseHTML(string(body))
}

// ownerRepoRe matches owner/repo patterns in skill entries.
var ownerRepoRe = regexp.MustCompile(`(?:href="[^"]*github\.com/|>)([a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+)`)

func parseHTML(html string) ([]RepoEntry, error) {
	type key struct{ owner, repo string }
	m := map[key]*RepoEntry{}

	matches := ownerRepoRe.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		ownerRepo := match[1]
		parts := strings.SplitN(ownerRepo, "/", 2)
		if len(parts) != 2 {
			continue
		}
		owner, repo := parts[0], parts[1]

		// Skip non-repo patterns
		if owner == "skills" || owner == "api" || repo == "skills" && owner == "skills" {
			continue
		}

		k := key{owner, repo}
		if e, ok := m[k]; ok {
			e.SkillCount++
		} else {
			m[k] = &RepoEntry{
				Owner:      owner,
				Repo:       repo,
				SkillCount: 1,
			}
		}
	}

	var result []RepoEntry
	for _, e := range m {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SkillCount > result[j].SkillCount
	})
	return result, nil
}
