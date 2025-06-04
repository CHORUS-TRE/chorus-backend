//go:build unit || integration || acceptance

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/store/postgres"
	"github.com/CHORUS-TRE/chorus-backend/tests/helpers"
	_ "github.com/lib/pq"
)

var (
	HarborRegistry = ""
	Username       = "" // From env var
	Password       = "" // From env var
	Project        = "apps"
	TenantID       = 1
	UserID         = 1
)

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

func getRepositories() ([]*Repository, error) {
	url := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories?page_size=100", HarborRegistry, Project)
	fmt.Println("url", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repositories: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	if err := json.Unmarshal(body, &repositories); err != nil {
		return nil, err
	}

	for _, repo := range repositories {
		if len(repo.Name) > 5 && repo.Name[:5] == fmt.Sprintf("%s/", Project) {
			repo.Name = repo.Name[5:]
		}
	}

	return repositories, nil
}

func getTags(repoName string) ([]string, error) {
	url := fmt.Sprintf("https://%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100", HarborRegistry, Project, repoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tags for repository %s: %s", repoName, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
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

// need to set env TEST_CONFIG_FILE with "./configs/dev/chorus.yaml" if run from base
// i.e. TEST_CONFIG_FILE="./configs/dev/chorus.yaml" go run --tags=unit ./tests/steward/copy-apps/main.go
func main() {
	helpers.Setup()

	// Override Harbor configuration with environment variables if provided
	if HarborRegistry_ := os.Getenv("CHORUS_HARBOR_URL"); HarborRegistry_ != "" {
		HarborRegistry = HarborRegistry_
	}
	if Project_ := os.Getenv("CHORUS_HARBOR_PROJECT"); Project_ != "" {
		Project = Project_
	}
	if Username_ := os.Getenv("CHORUS_HARBOR_USERNAME"); Username_ != "" {
		Username = Username_
	}
	if Password_ := os.Getenv("CHORUS_HARBOR_PASSWORD"); Password_ != "" {
		fmt.Printf("Using password from environment variable\n")
		Password = Password_
	}

	// Initialize configuration
	db := provider.ProvideMainDB(provider.WithClient("app-store"))
	if db == nil {
		fmt.Println("Error: Could not initialize database connection")
		return
	}

	// Create app storage
	appStore := postgres.NewAppStorage(db.DB.GetSqlxDB())

	// Get repositories from Harbor
	repos, err := getRepositories()
	if err != nil {
		fmt.Printf("Error fetching repositories: %v\n", err)
		return
	}

	// Process each repository
	for _, repo := range repos {
		fmt.Printf("Processing repository: %s\n", repo.Name)

		// Get tags for the repository
		tags, err := getTags(repo.Name)
		if err != nil {
			fmt.Printf("Error fetching tags for repository %s: %v\n", repo.Name, err)
			continue
		}

		if len(tags) == 0 {
			fmt.Printf("No tags found for repository %s, skipping...\n", repo.Name)
			continue
		}

		// Create app in database
		_, err = appStore.CreateApp(context.Background(), uint64(TenantID), &model.App{
			UserID:              uint64(UserID),
			Name:                repo.Name,
			Description:         repo.Name,
			Status:              model.AppActive,
			DockerImageRegistry: fmt.Sprintf("%s/%s", HarborRegistry, Project),
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
