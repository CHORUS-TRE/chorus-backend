package harbor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type App struct {
	Repository string            `json:"repository"`
	Tags       []string          `json:"tags"`
	Labels     map[string]string `json:"labels"`
}

type HarborClient interface {
	ListApps() ([]App, error)
}

type HarborNoopClient struct{}

func (c *HarborNoopClient) ListApps() ([]App, error) {
	return nil, nil
}

type harborClient struct {
	cfg    config.HarborClient
	client *http.Client
}

func NewHarborClient(cfg config.Config) HarborClient {
	return &harborClient{
		cfg: cfg.Clients.HarborClient,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
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

// registryManifest is a Docker/OCI manifest (schema v2).
type registryManifest struct {
	Config manifestDescriptor `json:"config"`
}

type manifestDescriptor struct {
	Digest string `json:"digest"`
}

// imageConfig is the OCI image config JSON.
type imageConfig struct {
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"config"`
}

func (c *harborClient) ListApps() ([]App, error) {
	repos, err := c.listRepositories()
	if err != nil {
		return nil, fmt.Errorf("listing repositories: %w", err)
	}

	var apps []App
	for _, repo := range repos {
		name := c.stripProjectPrefix(repo.Name)
		artifacts, err := c.listArtifacts(name)
		if err != nil {
			return nil, fmt.Errorf("listing artifacts for %s: %w", name, err)
		}

		for _, artifact := range artifacts {
			tags := extractTagNames(artifact.Tags)
			if len(tags) == 0 {
				continue
			}

			labels, err := c.fetchLabels(name, artifact.Digest)
			if err != nil {
				return nil, fmt.Errorf("fetching labels for %s@%s: %w", name, artifact.Digest, err)
			}

			apps = append(apps, App{
				Repository: name,
				Tags:       tags,
				Labels:     labels,
			})
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

// fetchLabels retrieves an artifact's OCI image config via the Docker Registry
// HTTP V2 API and returns the labels that match the configured prefixes.
func (c *harborClient) fetchLabels(repoName, digest string) (map[string]string, error) {
	manifest, err := c.fetchManifest(repoName, digest)
	if err != nil {
		return nil, err
	}

	imgCfg, err := c.fetchImageConfig(repoName, manifest.Config.Digest)
	if err != nil {
		return nil, err
	}

	return c.filterLabels(imgCfg.Config.Labels), nil
}

func (c *harborClient) fetchManifest(repoName, reference string) (*registryManifest, error) {
	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s",
		c.cfg.URL, c.cfg.Project, repoName, reference)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
	c.setAuth(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching manifest %s@%s: status %d", repoName, reference, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m registryManifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("decoding manifest: %w", err)
	}
	return &m, nil
}

func (c *harborClient) fetchImageConfig(repoName, digest string) (*imageConfig, error) {
	url := fmt.Sprintf("%s/v2/%s/%s/blobs/%s",
		c.cfg.URL, c.cfg.Project, repoName, digest)

	body, err := c.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("fetching image config: %w", err)
	}

	var cfg imageConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("decoding image config: %w", err)
	}
	return &cfg, nil
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
	if c.cfg.Username != "" && c.cfg.Password != "" {
		req.SetBasicAuth(c.cfg.Username, string(c.cfg.Password))
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

func extractTagNames(tags []harborTag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}
