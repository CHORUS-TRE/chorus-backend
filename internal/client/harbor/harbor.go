package harbor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/docker"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type App struct {
	Repository string            `json:"repository"`
	Tag        string            `json:"tag"`
	Labels     map[string]string `json:"labels"`
}

type HarborClient interface {
	ListApps(existingApps map[string]struct{}) ([]App, error)
}

type HarborNoopClient struct{}

func (c *HarborNoopClient) ListApps(existingApps map[string]struct{}) ([]App, error) {
	return nil, nil
}

type harborClient struct {
	cfg          config.HarborClient
	client       *http.Client
	dockerClient docker.DockerClienter
	registryHost string
}

func NewHarborClient(cfg config.Config, dockerClient docker.DockerClienter) HarborClient {
	harborCfg := cfg.Clients.HarborClient

	registryHost := ""
	if u, err := url.Parse(harborCfg.URL); err == nil {
		registryHost = u.Host
	}

	return &harborClient{
		cfg: harborCfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		dockerClient: dockerClient,
		registryHost: registryHost,
	}
}

// harborRepository is the Harbor API response for a repository entry.
type harborRepository struct {
	Name string `json:"name"`
}

// harborArtifact is the Harbor API response for an artifact entry.
type harborArtifact struct {
	Digest string      `json:"digest"`
	Tags   []harborTag `json:"tags"`
}

type harborTag struct {
	Name string `json:"name"`
}

func (c *harborClient) ListApps(existingApps map[string]struct{}) ([]App, error) {
	repos, err := c.listRepositories()
	if err != nil {
		return nil, fmt.Errorf("listing repositories: %w", err)
	}

	var apps []App
	for _, repo := range repos {
		strippedName := c.stripProjectPrefix(repo.Name)
		artifacts, err := c.listArtifacts(strippedName)
		if err != nil {
			return nil, fmt.Errorf("listing artifacts for %s: %w", strippedName, err)
		}

		for _, artifact := range artifacts {
			tags := extractTagNames(artifact.Tags)
			if len(tags) == 0 {
				continue
			}

			if allTagsExist(existingApps, repo.Name, tags) {
				continue
			}

			labels, err := c.fetchLabels(strippedName, artifact.Digest)
			if err != nil {
				return nil, fmt.Errorf("fetching labels for %s@%s: %w", strippedName, artifact.Digest, err)
			}

			for _, tag := range tags {
				apps = append(apps, App{
					Repository: repo.Name,
					Tag:        tag,
					Labels:     labels,
				})
			}
		}
	}

	return apps, nil
}

func (c *harborClient) listRepositories() ([]harborRepository, error) {
	var allRepos []harborRepository
	pageSize := c.pageSize()

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page_size=%d&page=%d",
			c.cfg.URL, c.cfg.Project, pageSize, page)

		body, err := c.doGet(url)
		if err != nil {
			return nil, err
		}

		var repos []harborRepository
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("decoding repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if len(repos) < pageSize {
			break
		}
	}

	return allRepos, nil
}

func (c *harborClient) listArtifacts(repoName string) ([]harborArtifact, error) {
	var allArtifacts []harborArtifact
	pageSize := c.pageSize()

	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=%d&page=%d",
			c.cfg.URL, c.cfg.Project, repoName, pageSize, page)

		body, err := c.doGet(url)
		if err != nil {
			return nil, err
		}

		var artifacts []harborArtifact
		if err := json.Unmarshal(body, &artifacts); err != nil {
			return nil, fmt.Errorf("decoding artifacts: %w", err)
		}

		allArtifacts = append(allArtifacts, artifacts...)

		if len(artifacts) < pageSize {
			break
		}
	}

	return allArtifacts, nil
}

// fetchLabels builds a full image reference and delegates to the Docker client
// to retrieve OCI image config labels, then filters by configured prefixes.
func (c *harborClient) fetchLabels(repoName, digest string) (map[string]string, error) {
	imageRef := fmt.Sprintf("%s/%s/%s@%s", c.registryHost, c.cfg.Project, repoName, digest)

	allLabels, err := c.dockerClient.GetLabels(imageRef, c.cfg.Username, c.cfg.Password.PlainText())
	if err != nil {
		return nil, fmt.Errorf("getting labels for %s: %w", imageRef, err)
	}

	return c.filterLabels(allLabels), nil
}

func (c *harborClient) filterLabels(all map[string]string) map[string]string {
	if len(c.cfg.LabelPrefixes) == 0 {
		return all
	}

	filtered := make(map[string]string, len(all))
	for k, v := range all {
		for _, prefix := range c.cfg.LabelPrefixes {
			if strings.HasPrefix(k, prefix) {
				filtered[k] = v
				break
			}
		}
	}
	return filtered
}

func (c *harborClient) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *harborClient) setAuth(req *http.Request) {
	if c.cfg.Username != "" && c.cfg.Password.PlainText() != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password.PlainText())
	}
}

func (c *harborClient) stripProjectPrefix(name string) string {
	prefix := c.cfg.Project + "/"
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix):]
	}
	return name
}

func (c *harborClient) pageSize() int {
	if c.cfg.PageSize > 0 {
		return c.cfg.PageSize
	}
	return 100
}

func allTagsExist(existing map[string]struct{}, repo string, tags []string) bool {
	if len(existing) == 0 {
		return false
	}
	for _, tag := range tags {
		if _, ok := existing[repo+":"+tag]; !ok {
			return false
		}
	}
	return true
}

func extractTagNames(tags []harborTag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}
