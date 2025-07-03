package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	configDevPath         = "./configs/dev"
	configDevOverridePath = "./configs/dev/files/kind.yaml"
	configDevFilename     = "chorus"
)

type HarborConfig struct {
	Registry string
	Project  string
	Username string
	Password string
	TenantID int64
	UserID   int64
}

type Repository struct {
	Name string `json:"name"`
}

type Artifact struct {
	Type string `json:"type"`
	Tags []*Tag `json:"tags"`
}

type Tag struct {
	Name string `json:"name"`
}

func promptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (harborCfg *HarborConfig) performRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(harborCfg.Username, harborCfg.Password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (harborCfg *HarborConfig) getRepositories() ([]*Repository, error) {
	url := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories?page_size=100", harborCfg.Registry, harborCfg.Project)
	body, err := harborCfg.performRequest(url)
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	if err := json.Unmarshal(body, &repositories); err != nil {
		return nil, err
	}

	for _, repo := range repositories {
		if len(repo.Name) > 5 && repo.Name[:5] == fmt.Sprintf("%s/", harborCfg.Project) {
			repo.Name = repo.Name[5:]
		}
	}

	return repositories, nil
}

func (harborCfg *HarborConfig) getTags(repoName string) ([]string, error) {
	url := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100", harborCfg.Registry, harborCfg.Project, repoName)
	body, err := harborCfg.performRequest(url)
	if err != nil {
		return nil, err
	}

	var artifacts []*Artifact
	if err := json.Unmarshal(body, &artifacts); err != nil {
		return nil, err
	}

	var tags []string
	for _, artifact := range artifacts {
		for _, tag := range artifact.Tags {
			tags = append(tags, tag.Name)
		}
	}

	return tags, nil
}

func main() {
	defer logPanicRecovery()

	cfg := provider.ProvideConfig()

	err := runExportConfig()

	// Initialize loggers
	stopLoggers, err := logger.InitLoggers(cfg)
	if err != nil {
		fmt.Printf("failed to initialize loggers: %v\n", err)
		return
	}
	defer stopLoggers()

	// Initialize provider
	appStore := provider.ProvideAppStore()
	if appStore == nil {
		fmt.Println("Failed to initialize app store")
		return
	}

	// Configure Harbor connection
	harborCfg := &HarborConfig{
		Registry: promptInput("Enter Harbor registry URL: "),
		Username: promptInput("Enter Harbor username: "),
		Password: promptInput("Enter Harbor password: "),
		Project:  "apps",
		TenantID: 1,
		UserID:   1,
	}

	// Get repositories from Harbor
	repos, err := harborCfg.getRepositories()
	if err != nil {
		fmt.Printf("Error fetching repositories: %v\n", err)
		return
	}

	// Process each repository
	for _, repo := range repos {
		fmt.Printf("Processing repository: %s\n", repo.Name)

		// Get tags for the repository
		tags, err := harborCfg.getTags(repo.Name)
		if err != nil {
			fmt.Printf("Error fetching tags for repository %s: %v\n", repo.Name, err)
			continue
		}

		if len(tags) == 0 {
			fmt.Printf("No tags found for repository %s, skipping...\n", repo.Name)
			continue
		}

		// Create app in database
		_, err = appStore.CreateApp(context.Background(), uint64(harborCfg.TenantID), &model.App{
			UserID:              uint64(harborCfg.UserID),
			Name:                repo.Name,
			Description:         repo.Name,
			Status:              model.AppActive,
			DockerImageRegistry: fmt.Sprintf("%s/%s", harborCfg.Registry, harborCfg.Project),
			DockerImageName:     repo.Name,
			DockerImageTag:      tags[0],
		})

		if err != nil {
			fmt.Printf("Error creating app %s: %v\n", repo.Name, err)
			continue
		}
		fmt.Printf("Successfully created app %s with tag %s\n", repo.Name, tags[0])
	}
}

func logPanicRecovery() {
	if r := recover(); r != nil {
		logger.TechLog.Fatal(context.Background(), "goodbye world, panic occurred", zap.String("panic_error", fmt.Sprintf("%v", r)), zap.Stack("panic_stack_trace"))
	}
}

// runExportConfig outputs the loaded configuration to stdout.
func runExportConfig() error {
	cfg := provider.ProvideConfig()

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}
